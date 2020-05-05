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
	entry, err := c.validate(ctx, req)
	if err != nil {
		logrus.WithError(err).Error("failed to validate entry")
		return common.Response(ctx, http.StatusBadRequest, nil), nil
	}

	if _, err := c.DynamoDB.PutItem(entry.PutItem()); err != nil {
		logrus.WithError(err).Error("failed to put item")
		return common.Response(ctx, http.StatusOK, nil), nil
	}

	logrus.WithFields(lf).Infof("successfully handled twilio webhook")
	return common.Response(ctx, http.StatusOK, entry), nil
}

type TwilioInboundReq struct {
	From string `schema:"From"`
	Body string `schema:"Body"`
}

func (c controller) validate(ctx context.Context, req *events.APIGatewayProxyRequest) (journal.Entry, error) {
	e := journal.Entry{
		EntryID: "entry_" + uuid.New().String(),
		Created: time.Now().Format(time.RFC3339),
		Updated: time.Now().Format(time.RFC3339),
	}
	logrus.Infof("REQBODY %s", req.Body)
	params, err := url.ParseQuery(req.Body)
	if err != nil {
		return e, errors.Wrap(err, "failed to parse query")
	}
	logrus.Infof("params %+v", params)
	payload := TwilioInboundReq{}
	decoder := schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)
	if err := decoder.Decode(&payload, params); err != nil {
		return e, errors.Wrap(err, "failed to decode twilio payload")
	}
	logrus.Infof("payload %+v", payload)
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
