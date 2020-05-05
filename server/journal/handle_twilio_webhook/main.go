package main

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/kelseyhightower/envconfig"
	"github.com/mikerjacobi/lambda-journal/server/common"
	"github.com/sirupsen/logrus"
)

type controller struct {
	common.Controller
	Test string `envconfig:"JOURNAL_TEST"`
}

func (c controller) handleTwilioWebhook(ctx context.Context, req *events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	lf := logrus.Fields{"handler": "handle_twilio_webhook"}

	sess := session.Must(session.NewSession(&aws.Config{
		Endpoint: aws.String("http://dev.jaqobi.com:8000"),
		Region:   aws.String("us-west-2"),
	}))

	// Create DynamoDB client
	svc := dynamodb.New(sess)
	result, err := svc.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String("usersTable"),
		Item: map[string]*dynamodb.AttributeValue{
			"email": {S: aws.String("bob@lob.cob")},
			"name": {
				S: aws.String("bob"),
			},
		},
	})
	if err != nil {
		logrus.WithError(err).Error("failed to put item")
		return common.Response(ctx, http.StatusOK, nil), nil
	}

	logrus.WithFields(lf).Infof("successfully handled twilio webhook %v", result)
	return common.Response(ctx, http.StatusOK, nil), nil
}

func main() {
	service := "handle-twilio-webhook"
	c := controller{Controller: common.Controller{}}
	if err := envconfig.Process("journal", &c); err != nil {
		logrus.WithError(err).Panicf("failed to load config")
	}

	if c.Environment == "sandbox" {
		c.Controller.ServeSandbox("/twilio", "POST", c.handleTwilioWebhook)
	} else {
		logrus.Infof("starting lambda %s server", service)
		lambda.Start(c.handleTwilioWebhook)
	}
}
