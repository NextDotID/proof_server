/// Lambda worker for revalidating proof records
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	myconfig "github.com/nextdotid/proof-server/config"
	"github.com/nextdotid/proof-server/model"
	"github.com/nextdotid/proof-server/types"
	"github.com/nextdotid/proof-server/validator/das"
	"github.com/nextdotid/proof-server/validator/discord"
	"github.com/nextdotid/proof-server/validator/ethereum"
	"github.com/nextdotid/proof-server/validator/github"
	"github.com/nextdotid/proof-server/validator/keybase"
	"github.com/nextdotid/proof-server/validator/solana"
	"github.com/nextdotid/proof-server/validator/twitter"
	"github.com/sirupsen/logrus"
	"golang.org/x/xerrors"
)

var initialized = false

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
		if message.Action != types.QueueActions.Revalidate {
			continue
		}
		if err := do_single(ctx, &message); err != nil {
			return err
		}
	}
	return nil
}

func do_single(ctx context.Context, message *types.QueueMessage) error {
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
