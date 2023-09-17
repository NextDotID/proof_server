// / Lambda worker for revalidating proof records
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/everFinance/goar"
	artypes "github.com/everFinance/goar/types"
	"github.com/everFinance/goar/utils"
	"github.com/nextdotid/proof_server/common"
	myconfig "github.com/nextdotid/proof_server/config"
	"github.com/nextdotid/proof_server/model"
	"github.com/nextdotid/proof_server/types"
	"github.com/nextdotid/proof_server/util"
	utilS3 "github.com/nextdotid/proof_server/util/s3"
	"github.com/nextdotid/proof_server/validator/activitypub"
	"github.com/nextdotid/proof_server/validator/das"
	"github.com/nextdotid/proof_server/validator/discord"
	"github.com/nextdotid/proof_server/validator/dns"
	"github.com/nextdotid/proof_server/validator/ethereum"
	"github.com/nextdotid/proof_server/validator/github"
	"github.com/nextdotid/proof_server/validator/keybase"
	"github.com/nextdotid/proof_server/validator/minds"
	"github.com/nextdotid/proof_server/validator/solana"
	"github.com/nextdotid/proof_server/validator/steam"
	"github.com/nextdotid/proof_server/validator/twitter"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"golang.org/x/xerrors"
	"gorm.io/gorm/clause"
)

var (
	awsConfig   aws.Config
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

func handler(ctx context.Context, sqs_event events.SQSEvent) (events.SQSEventResponse, error) {
	failures := []events.SQSBatchItemFailure{}
	// Key: persona, Value: message id; persona as key to uniq
	arweaveMsgs := map[string]string{}
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
			fmt.Printf("error deserializing sqs records: %+v\n", raw_message.Body)
			failures = append(failures, events.SQSBatchItemFailure{ItemIdentifier: raw_message.MessageId})
		}

		switch message.Action {
		case types.QueueActions.ArweaveUpload:
			arweaveMsgs[message.Persona] = raw_message.MessageId
		case types.QueueActions.Revalidate:
			if err := revalidateSingle(ctx, &message); err != nil {
				fmt.Printf("error revalidating proof record %d: %s\n", message.ProofID, err)
				// Ignore failed revalidation job since failed job will still update DB.
				// failures = append(failures, events.SQSBatchItemFailure{ItemIdentifier: raw_message.MessageId})
			}
		case types.QueueActions.TwitterOAuthTokenAcquire:
			{
				err := twitterRefreshOAuthToken()
				if err != nil {
					// Ignore errors for now
					fmt.Printf("Error when retrieving Twitter OAuth key: %s", err.Error())
				}
			}
		default:
			logrus.Warnf("unsupported queue action: %s", message.Action)
			failures = append(failures, events.SQSBatchItemFailure{ItemIdentifier: raw_message.MessageId})
		}
	}

	arweaveFailed := lo.MapToSlice(arweaveMsgs, func(persona string, id string) events.SQSBatchItemFailure {
		return events.SQSBatchItemFailure{ItemIdentifier: id}
	})

	// At least we need 5 persona in a batch to be cost-effective
	if len(arweaveMsgs) < MIN_PERSONAS {
		failures = append(failures, arweaveFailed...)
	} else if err := arweave_upload_many(lo.Keys(arweaveMsgs)); err != nil {
		failures = append(failures, arweaveFailed...)
	}

	return events.SQSEventResponse{BatchItemFailures: failures}, nil
}

func arweave_upload_many(personas []string) error {
	if wallet == nil {
		return xerrors.New("wallet is not initialized")
	}

	tx := model.DB.Begin()
	items := []artypes.BundleItem{}

	for _, persona := range personas {
		chains := []*model.ProofChain{}
		findTx := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("persona = ?", persona).
			Order("ID asc").
			Find(&chains)
		if findTx.Error != nil {
			tx.Rollback()
			return xerrors.Errorf("error when find and lock proof chains: %w", findTx.Error)
		}

		if len(chains) == 0 {
			logrus.Warnf("no chains to upload: %s", persona)
			return nil
		}

		for _, pc := range chains {
			if pc.ArweaveID != "" {

			}

			previous, ok := lo.Find(chains, func(item *model.ProofChain) bool {
				return pc.PreviousID.Valid && pc.PreviousID.Int64 == item.ID
			})
			if ok && previous.ArweaveID == "" {
				logrus.Warnf("previous chain is not uploaded yet: %d", previous.ID)
				break
			}

			item, err := arweaveBundleSingle(pc, previous)
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

func arweaveBundleSingle(pc *model.ProofChain, previous *model.ProofChain) (*artypes.BundleItem, error) {
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
		AltID:             pc.AltID,
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

func revalidateSingle(ctx context.Context, message *types.QueueMessage) error {
	proof := model.Proof{}
	tx := model.DB.Preload("ProofChain").Preload("ProofChain.Previous").Where("id = ?", message.ProofID).First(&proof)
	if tx.Error != nil {
		return xerrors.Errorf("%w", tx.Error)
	}
	return proof.Revalidate()
}

func initDB() {
	shouldMigrate := util.GetE("DB_MIGRATE", "false")
	model.Init(shouldMigrate == "true")
}

// func init_sqs(cfg aws.Config) {
// 	sqs.Init(cfg)
// }

func initValidators() {
	twitter.Init()
	ethereum.Init()
	keybase.Init()
	github.Init()
	discord.Init()
	das.Init()
	solana.Init()
	minds.Init()
	dns.Init()
	steam.Init()
	activitypub.Init()
}

func init() {
	var err error
	awsConfig, err = config.LoadDefaultConfig(
		context.Background(),
		config.WithRegion("ap-east-1"),
	)
	if err != nil {
		logrus.Fatalf("Unable to load AWS config: %s", err)
	}
	common.CurrentRuntime = common.Runtimes.Lambda
	initConfigFromAWSSecret()
	logrus.SetLevel(logrus.InfoLevel)

	initDB()
	// init_sqs(cfg)
	initValidators()
}

func initConfigFromAWSSecret() {
	if initialized {
		return
	}
	secretName := util.GetE("SECRET_NAME", "")
	region := util.GetE("SECRET_REGION", "")

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
		SecretId:     aws.String(secretName),
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

func twitterRefreshOAuthToken() (err error) {
	const VALID_TOKEN_AMOUNT = 5
	if err = utilS3.Init(awsConfig); err != nil {
		logrus.Fatalf("Error during initializing S3 client: %s", err.Error())
	}
	ctx := context.Background()
	tokens, err := twitter.GetTokenListFromS3(ctx)
	if err != nil {
		logrus.Fatalf("Error when loading Twitter token list from S3: %s", err.Error())
	}

	validTokens := lo.Filter(tokens.Tokens, func(token twitter.Token, _index int) bool {
		return !token.IsExpired()
	})
	if len(validTokens) < VALID_TOKEN_AMOUNT {
		// Generate a new one
		newToken, err := twitter.GenerateOauthToken()
		if err != nil {
			return err
		}
		fmt.Printf("TWITTER OAUTH KEY REGISTERED: %+v", *tokens)
		validTokens = append(validTokens, *newToken)
		newTokenList := twitter.TokenList{
			Tokens: validTokens,
		}
		newTokenListJSON, _ := newTokenList.ToJSON()
		utilS3.PutToS3(ctx, twitter.TWITTER_TOKEN_LIST_FILENAME, newTokenListJSON)
	}

	return nil
}
