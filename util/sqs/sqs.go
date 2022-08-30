package sqs

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	myconfig "github.com/nextdotid/proof_server/config"
	"github.com/sirupsen/logrus"
	"golang.org/x/xerrors"
)

var (
	api      SQSSendMessageAPI
	queueUrl *string
)

type SQSSendMessageAPI interface {
	GetQueueUrl(ctx context.Context,
		params *sqs.GetQueueUrlInput,
		optFns ...func(*sqs.Options)) (*sqs.GetQueueUrlOutput, error)

	SendMessage(ctx context.Context,
		params *sqs.SendMessageInput,
		optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error)
}

func getQueueUrl(c context.Context, api SQSSendMessageAPI, input *sqs.GetQueueUrlInput) (*sqs.GetQueueUrlOutput, error) {
	return api.GetQueueUrl(c, input)
}

func sendMessage(c context.Context, api SQSSendMessageAPI, input *sqs.SendMessageInput) (*sqs.SendMessageOutput, error) {
	return api.SendMessage(c, input)
}

func Init(cfg aws.Config) {
	queueName := myconfig.C.Sqs.QueueName
	if queueName == "" {
		logrus.Fatal("queue_name is not set")
	}

	api = sqs.NewFromConfig(cfg)

	gqInput := &sqs.GetQueueUrlInput{QueueName: aws.String(queueName)}
	result, err := getQueueUrl(context.TODO(), api, gqInput)
	if err != nil {
		logrus.Fatalf("error getting queue url: %v", err)
	}

	queueUrl = result.QueueUrl
}

func Send(msg interface{}) error {
	if api == nil {
		return xerrors.New("sqs is not initialized")
	}

	json, err := json.Marshal(msg)
	if err != nil {
		logrus.Errorf("error marshalling message: %v", err)
		return err
	}

	smInput := &sqs.SendMessageInput{
		MessageBody: aws.String(string(json)),
		QueueUrl:    queueUrl,
	}

	_, err = sendMessage(context.TODO(), api, smInput)
	if err != nil {
		logrus.Errorf("error sending message to sqs: %v", err)
		return err
	}

	return nil
}
