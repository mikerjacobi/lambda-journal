package main

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/google/uuid"
	"github.com/gorilla/schema"
	"github.com/kelseyhightower/envconfig"
	"github.com/mikerjacobi/lambda-journal/server/common"
	"github.com/mikerjacobi/lambda-journal/server/journal"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type controller struct {
	common.Controller
	Test string `envconfig:"JOURNAL_TEST"`
}

func (c controller) handleTwilioWebhook(ctx context.Context, req *events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	lf := logrus.Fields{"handler": c.Service}
	ctx = common.WithContentType(ctx, "text/xml")

	journal, err := c.validate(ctx, req)
	if err != nil {
		logrus.WithError(err).Error("failed to validate journal")
		return common.Response(ctx, http.StatusOK, nil), nil
	}

	if _, err := c.DynamoDB.PutItem(journal.PutItem()); err != nil {
		logrus.WithError(err).Error("failed to put item")
		return common.Response(ctx, http.StatusOK, nil), nil
	}

	logrus.WithFields(lf).Infof("successfully handled twilio webhook")
	return common.Response(ctx, http.StatusOK, ""), nil
}

type TwilioInboundReq struct {
	From string `schema:"From"`
	Body string `schema:"Body"`
}

func (c controller) validate(ctx context.Context, req *events.APIGatewayProxyRequest) (journal.Journal, error) {
	e := journal.Journal{
		JournalID: "journal:" + uuid.New().String(),
		Created:   time.Now().Format(time.RFC3339),
		Updated:   time.Now().Format(time.RFC3339),
	}
	params, err := url.ParseQuery(req.Body)
	if err != nil {
		return e, errors.Wrap(err, "failed to parse query")
	}
	payload := TwilioInboundReq{}
	decoder := schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)
	if err := decoder.Decode(&payload, params); err != nil {
		return e, errors.Wrap(err, "failed to decode twilio payload")
	}
	e.Entry = payload.Body
	return e, nil
}

func main() {
	c := controller{Controller: common.Controller{Service: "handle-twilio-webhook"}}
	if err := envconfig.Process("journal", &c); err != nil {
		logrus.WithError(err).Panicf("failed to load config")
	}
	common.Init(&c.Controller)

	if c.Environment == "sandbox" {
		c.Controller.ServeSandbox("/twilio", "POST", c.handleTwilioWebhook)
	} else {
		lambda.Start(c.handleTwilioWebhook)
	}
}
