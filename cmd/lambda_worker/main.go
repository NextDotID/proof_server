/// Lambda worker for revalidating proof records
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
	myconfig "github.com/nextdotid/proof-server/config"
	"github.com/nextdotid/proof-server/model"
	"github.com/nextdotid/proof-server/types"
	"github.com/nextdotid/proof-server/validator/das"
	"github.com/nextdotid/proof-server/validator/discord"
	"github.com/nextdotid/proof-server/validator/ethereum"
	"github.com/nextdotid/proof-server/validator/github"
	"github.com/nextdotid/proof-server/validator/keybase"
	"github.com/nextdotid/proof-server/validator/twitter"
	"github.com/sirupsen/logrus"
	"golang.org/x/xerrors"
	"gorm.io/gorm/clause"
)

var (
	initialized = false
	wallet      *goar.Wallet
)

func main() {
	lambda.Start(handler)
}

func handler(ctx context.Context, sqs_event events.SQSEvent) error {
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
			if err := arweave_upload_many(&message); err != nil {
				return err
			}
		case types.QueueActions.Revalidate:
			if err := revalidate_single(ctx, &message); err != nil {
				return err
			}
		default:
			logrus.Warnf("unsupported queue action: %s", message.Action)
		}
	}

	return nil
}

func arweave_upload_many(message *types.QueueMessage) error {
	if wallet == nil {
		return xerrors.New("wallet is not initialized")
	}

	tx := model.DB.Begin()

	chains := []model.ProofChain{}
	err := tx.
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("persona = ? AND arweave_id = ?", message.Persona, "").
		Order("ID asc").
		Preload("Previous").
		Find(&chains)
	if err != nil {
		tx.Rollback()
		return xerrors.Errorf("error when find and lock proof chains: %w", err)
	}

	if len(chains) == 0 {
		// No chains to upload.
		return nil
	}

	for _, pc := range chains {
		if pc.ArweaveID != "" || (pc.PreviousID.Valid && pc.Previous.ArweaveID == "") {
			continue
		}

		if err := arweave_upload_single(&pc); err != nil {
			logrus.Errorf("error uploading proof chain %s: %w", pc.Uuid, err)
			break
		}
	}

	if err = tx.Save(&chains); err != nil {
		tx.Rollback()
		return xerrors.Errorf("error saving proof chains arweave id updates: %w", err)
	}

	if commiTx := tx.Commit(); commiTx.Error != nil {
		return xerrors.Errorf("error committing transaction: %w", commiTx.Error)
	}

	return nil
}

func arweave_upload_single(pc *model.ProofChain) error {
	doc := model.ProofChainArweaveDocument{
		Action:            pc.Action,
		Platform:          pc.Platform,
		Identity:          pc.Identity,
		ProofLocation:     pc.Location,
		CreatedAt:         strconv.FormatInt(pc.CreatedAt.Unix(), 10),
		Signature:         pc.Signature,
		SignaturePayload:  pc.SignaturePayload,
		Uuid:              pc.Uuid,
		Extra:             pc.Extra,
		PreviousUuid:      pc.Previous.Uuid,
		PreviousArweaveID: pc.Previous.ArweaveID,
	}

	json, err := json.MarshalIndent(doc, "", "\t")
	if err != nil {
		return xerrors.Errorf("error marshalling document: %w", err)
	}

	artx, err := wallet.SendData(json, []artypes.Tag{
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
		return xerrors.Errorf("error sending data to arweave: %s, %w", artx.ID, err)
	}

	pc.ArweaveID = artx.ID

	return nil
}

func revalidate_single(ctx context.Context, message *types.QueueMessage) error {
	proof := model.Proof{}
	tx := model.DB.Preload("ProofChain.Previous").First(&proof, message.ProofID)
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
