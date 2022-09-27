package main

import (
	"context"
	"encoding/json"
	"os"

	"github.com/akrylysov/algnhsa"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	myconfig "github.com/nextdotid/proof_server/config"
	"github.com/nextdotid/proof_server/controller"
	"github.com/nextdotid/proof_server/model"
	"github.com/nextdotid/proof_server/util/sqs"
	"github.com/nextdotid/proof_server/validator/das"
	"github.com/nextdotid/proof_server/validator/discord"
	"github.com/nextdotid/proof_server/validator/dns"
	"github.com/nextdotid/proof_server/validator/ethereum"
	"github.com/nextdotid/proof_server/validator/github"
	"github.com/nextdotid/proof_server/validator/keybase"
	"github.com/nextdotid/proof_server/validator/minds"
	"github.com/nextdotid/proof_server/validator/solana"
	"github.com/nextdotid/proof_server/validator/twitter"
	"github.com/sirupsen/logrus"
)

var (
	initialized = false
)

func init_db(cfg aws.Config) {
	model.Init()
}

func init_sqs(cfg aws.Config) {
	sqs.Init(cfg)
}

func init_validators() {
	twitter.Init()
	ethereum.Init()
	keybase.Init()
	github.Init()
	discord.Init()
	das.Init()
	solana.Init()
	minds.Init()
	dns.Init()
}

func init() {
	cfg, err := config.LoadDefaultConfig(
		context.Background(),
		// TODO: change region
		config.WithRegion("ap-east-1"),
	)
	if err != nil {
		logrus.Fatalf("Unable to load AWS config: %s", err)
	}
	init_config_from_aws_secret()
	logrus.SetLevel(logrus.WarnLevel)

	init_db(cfg)
	init_sqs(cfg)
	init_validators()
	controller.Init()
}

func main() {
	algnhsa.ListenAndServe(controller.Engine, nil)
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
