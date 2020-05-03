package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/husobee/vestigo"
	"github.com/kelseyhightower/envconfig"
	"github.com/mikerjacobi/lambda-journal/server/common"
	"github.com/sirupsen/logrus"
)

type controller struct {
	common.Controller
	Test string `envconfig:"JOURNAL_TEST"`
}

func (c controller) sandboxHandleTwilioWebhook(rw http.ResponseWriter, req *http.Request) {
	body, _ := ioutil.ReadAll(req.Body)
	apigReq := &events.APIGatewayProxyRequest{Body: string(body)}
	resp, _ := c.handleTwilioWebhook(context.Background(), apigReq)
	rw.WriteHeader(resp.StatusCode)
	rw.Write([]byte(resp.Body))
}

func (c controller) handleTwilioWebhook(ctx context.Context, req *events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	lf := logrus.Fields{"handler": "handle_twilio_webhook"}

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// Create DynamoDB client
	svc := dynamodb.New(sess)
	result, err := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"Year": {
				N: aws.String(movieYear),
			},
			"Title": {
				S: aws.String(movieName),
			},
		},
	})
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	logrus.WithFields(lf).Infof("successfully handled twilio webhook")
	return common.Response(ctx, http.StatusOK, nil), nil
}

func main() {
	service := "handle-twilio-webhook"
	c := controller{Controller: common.Controller{}}
	if err := envconfig.Process("myapp", &c); err != nil {
		logrus.WithError(err).Panicf("failed to load config")
	}

	if c.Environment == "sandbox" {
		logrus.Infof("starting sandbox %s server %+v", service, c)
		router := vestigo.NewRouter()
		router.Post("/twilio", c.sandboxHandleTwilioWebhook)
		logrus.Fatal(http.ListenAndServe("0.0.0.0:80", router))
	} else {
		logrus.Infof("starting lambda %s server", service)
		lambda.Start(c.handleTwilioWebhook)
	}
}
