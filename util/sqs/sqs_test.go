package sqs

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/stretchr/testify/assert"

	myconfig "github.com/nextdotid/proof_server/config"
)

type SQSSendMessageImpl struct{}

func (dt SQSSendMessageImpl) GetQueueUrl(ctx context.Context,
	params *sqs.GetQueueUrlInput,
	optFns ...func(*sqs.Options)) (*sqs.GetQueueUrlOutput, error) {

	prefix := "https://sqs.REGION.amazonaws.com/ACCOUNT#/"

	output := &sqs.GetQueueUrlOutput{
		QueueUrl: aws.String(prefix + "aws-docs-example-queue-url1"),
	}

	return output, nil
}

func (dt SQSSendMessageImpl) SendMessage(ctx context.Context,
	params *sqs.SendMessageInput,
	optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error) {

	output := &sqs.SendMessageOutput{
		MessageId: aws.String("aws-docs-example-messageID"),
	}

	return output, nil
}

func before_each(t *testing.T) {
	myconfig.C = &myconfig.Config{}
	myconfig.C.Sqs.QueueName = "test"

	api = &SQSSendMessageImpl{}

	gqInput := &sqs.GetQueueUrlInput{QueueName: aws.String(myconfig.C.Sqs.QueueName)}
	result, err := getQueueUrl(context.Background(), api, gqInput)
	if err != nil {
		t.Fatalf("error getting queue url: %v", err)
	}

	queueUrl = result.QueueUrl
}

func Test_Send(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		before_each(t)
		assert.NoError(t, Send(struct{ x string }{x: "y"}))
	})
}
