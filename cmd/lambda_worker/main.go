// / Lambda worker for revalidating proof records
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/everFinance/goar"
	artypes "github.com/everFinance/goar/types"
	"github.com/everFinance/goar/utils"
	myconfig "github.com/nextdotid/proof_server/config"
	"github.com/nextdotid/proof_server/model"
	"github.com/nextdotid/proof_server/types"
	"github.com/nextdotid/proof_server/validator/das"
	"github.com/nextdotid/proof_server/validator/discord"
	"github.com/nextdotid/proof_server/validator/ethereum"
	"github.com/nextdotid/proof_server/validator/github"
	"github.com/nextdotid/proof_server/validator/keybase"
	"github.com/nextdotid/proof_server/validator/solana"
	"github.com/nextdotid/proof_server/validator/twitter"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"golang.org/x/xerrors"
	"gorm.io/gorm/clause"
)

var (
	initialized = false
	wallet      *goar.Wallet
)

const (
	// Min number of messages to process at once
	MIN_PERSONAS = 5
)

func main() {
	lambda.Start(handler)
}

func handler(ctx context.Context, sqs_event events.SQSEvent) error {
	arweaveMsgs := []*types.QueueMessage{}
	for _, raw_message := range sqs_event.Records {
		fmt.Printf(
			"[%s] Received from [%s]: %s \n",
			raw_message.MessageId,
			raw_message.EventSource,
			raw_message.Body,
		)

		message := types.QueueMessage{}
		err := json.Unmarshal([]byte(raw_message.Body), &message)
		if err != nil {
			return err
		}

		switch message.Action {
		case types.QueueActions.ArweaveUpload:
			arweaveMsgs = append(arweaveMsgs, &message)
		case types.QueueActions.Revalidate:
			if err := revalidate_single(ctx, &message); err != nil {
				fmt.Printf("error revalidating proof record %d: %s\n", message.ProofID, err)
			}
		default:
			logrus.Warnf("unsupported queue action: %s", message.Action)
		}
	}

	arweaveMsgs = lo.UniqBy(arweaveMsgs, func(msg *types.QueueMessage) string {
		return msg.Persona
	})
	if len(arweaveMsgs) > 0 {
		// At least we need 5 persona in a batch to be cost-effective
		if len(arweaveMsgs) < MIN_PERSONAS {
			return xerrors.Errorf("received less than minimum: %d", len(arweaveMsgs))
		}

		if err := arweave_upload_many(arweaveMsgs); err != nil {
			return err
		}
	}

	return nil
}

func arweave_upload_many(messages []*types.QueueMessage) error {
	if wallet == nil {
		return xerrors.New("wallet is not initialized")
	}

	tx := model.DB.Begin()
	items := []artypes.BundleItem{}

	for _, message := range messages {
		chains := []*model.ProofChain{}
		findTx := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("persona = ?", message.Persona).
			Order("ID asc").
			Find(&chains)
		if findTx.Error != nil {
			tx.Rollback()
			return xerrors.Errorf("error when find and lock proof chains: %w", findTx.Error)
		}

		if len(chains) == 0 {
			logrus.Warnf("no chains to upload: %s", message.Persona)
			return nil
		}

		for _, pc := range chains {
			if pc.ArweaveID != "" {
				continue
			}

			previous, ok := lo.Find(chains, func(item *model.ProofChain) bool {
				return pc.PreviousID.Valid && pc.PreviousID.Int64 == item.ID
			})
			if ok && previous.ArweaveID == "" {
				logrus.Warnf("previous chain is not uploaded yet: %d", previous.ID)
				break
			}

			item, err := arweave_bundle_single(pc, previous)
			if err != nil {
				logrus.Errorf("error marshalling proof chain %s: %w", pc.Uuid, err)
				break
			}

			if saveTx := tx.Save(pc); saveTx.Error != nil {
				tx.Rollback()
				return xerrors.Errorf("error when save proof chain: %w", saveTx.Error)
			}

			items = append(items, *item)
		}
	}

	bundle, err := utils.NewBundle(items...)
	if err != nil {
		return xerrors.Errorf("error creating bundle: %w", err)
	}

	arTx, err := wallet.SendBundleTx(bundle.BundleBinary, []artypes.Tag{})
	if err != nil {
		return xerrors.Errorf("error sending bundle: %s, %w", arTx.ID, err)
	}

	if commitTx := tx.Commit(); commitTx.Error != nil {
		return xerrors.Errorf("error committing transaction: %w", commitTx.Error)
	}

	return nil
}

func arweave_bundle_single(pc *model.ProofChain, previous *model.ProofChain) (*artypes.BundleItem, error) {
	previousUuid := ""
	previousArweaveID := ""
	if previous != nil {
		previousUuid = previous.Uuid
		previousArweaveID = previous.ArweaveID
	}

	doc := model.ProofChainArweaveDocument{
		Avatar:            pc.Persona,
		Action:            pc.Action,
		Platform:          pc.Platform,
		Identity:          pc.Identity,
		ProofLocation:     pc.Location,
		CreatedAt:         strconv.FormatInt(pc.CreatedAt.Unix(), 10),
		Signature:         pc.Signature,
		SignaturePayload:  pc.SignaturePayload,
		Uuid:              pc.Uuid,
		Extra:             pc.Extra,
		PreviousUuid:      previousUuid,
		PreviousArweaveID: previousArweaveID,
	}

	json, err := json.MarshalIndent(doc, "", "\t")
	if err != nil {
		return nil, xerrors.Errorf("error marshalling document: %w", err)
	}

	item, err := wallet.CreateAndSignBundleItem(json, 1, "", "", []artypes.Tag{
		{
			Name:  "ProofService-UUID",
			Value: doc.Uuid,
		},
		{
			Name:  "ProofService-PreviousUUID",
			Value: doc.PreviousUuid,
		},
		{
			Name:  "ProofService-PreviousArweaveID",
			Value: doc.PreviousArweaveID,
		},
		{
			Name:  "Content-Type",
			Value: "application/json",
		},
	})
	if err != nil {
		return nil, xerrors.Errorf("error creating bundle item: %s, %w", item.Id, err)
	}

	pc.ArweaveID = item.Id
	return &item, nil
}

func revalidate_single(ctx context.Context, message *types.QueueMessage) error {
	proof := model.Proof{}
	tx := model.DB.Preload("ProofChain").Preload("ProofChain.Previous").Where("id = ?", message.ProofID).First(&proof)
	if tx.Error != nil {
		return xerrors.Errorf("%w", tx.Error)
	}
	_, err := proof.Revalidate()
	if err != nil {
		return xerrors.Errorf("%w", err)
	}
	return nil
}

func init_db(cfg aws.Config) {
	model.Init()
}

// func init_sqs(cfg aws.Config) {
// 	sqs.Init(cfg)
// }

func init_validators() {
	twitter.Init()
	ethereum.Init()
	keybase.Init()
	github.Init()
	discord.Init()
	das.Init()
	solana.Init()
}

func init() {
	cfg, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithRegion("ap-east-1"),
	)
	if err != nil {
		logrus.Fatalf("Unable to load AWS config: %s", err)
	}
	init_config_from_aws_secret()
	logrus.SetLevel(logrus.WarnLevel)

	init_db(cfg)
	// init_sqs(cfg)
	init_validators()
}

func init_config_from_aws_secret() {
	if initialized {
		return
	}
	secret_name := getE("SECRET_NAME", "")
	region := getE("SECRET_REGION", "")

	// Create a Secrets Manager client
	cfg, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithRegion(region),
	)
	if err != nil {
		logrus.Fatalf("Unable to load SDK config: %v", err)
	}

	client := secretsmanager.NewFromConfig(cfg)
	input := secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(secret_name),
		VersionStage: aws.String("AWSCURRENT"),
	}
	result, err := client.GetSecretValue(context.Background(), &input)
	if err != nil {
		logrus.Fatalf("Error occured: %s", err.Error())
	}

	// Decrypts secret using the associated KMS CMK.
	// Depending on whether the secret is a string or binary, one of these fields will be populated.
	if result.SecretString == nil {
		logrus.Fatalf("cannot get secret string")
	}
	secret_string := *result.SecretString

	err = json.Unmarshal([]byte(secret_string), myconfig.C)
	if err != nil {
		logrus.Fatalf("Error during parsing config JSON: %v", err)
	}

	// Arweave wallet initialize.
	wallet, err = goar.NewWallet([]byte(myconfig.C.Arweave.Jwk), myconfig.C.Arweave.ClientUrl)
	if err != nil {
		logrus.Fatalf("Error during Arweave wallet initialization: %v", err)
	}

	initialized = true
}

func getE(env_key, default_value string) string {
	result := os.Getenv(env_key)
	if len(result) == 0 {
		if len(default_value) > 0 {
			return default_value
		} else {
			logrus.Fatalf("ENV %s must be given! Abort.", env_key)
			return ""
		}

	} else {
		return result
	}
}
