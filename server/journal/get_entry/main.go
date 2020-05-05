package main

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/kelseyhightower/envconfig"
	"github.com/mikerjacobi/lambda-journal/server/common"
	"github.com/mikerjacobi/lambda-journal/server/journal"
	"github.com/sirupsen/logrus"
)

type controller struct {
	common.Controller
}

func (c controller) getEntry(ctx context.Context, req *events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	lf := logrus.Fields{"handler": "get_entry"}

	entry, err := c.validate(ctx, req)
	if err != nil {
		logrus.WithError(err).Error("failed to validate entry")
		return common.Response(ctx, http.StatusBadRequest, nil), nil
	}

	item, err := c.DynamoDB.GetItem(entry.GetItem())
	if err != nil {
		logrus.WithError(err).Error("failed to get item")
		return common.Response(ctx, http.StatusNotFound, nil), nil
	}
	entry = entry.GotItem(item)

	logrus.WithFields(lf).Infof("successfully got entry")
	return common.Response(ctx, http.StatusOK, entry), nil
}

func (c controller) validate(ctx context.Context, req *events.APIGatewayProxyRequest) (journal.Entry, error) {
	e := journal.Entry{EntryID: req.PathParameters["entry_id"]}
	res, err := common.GetResource(e.EntryID)
	if err != nil || res != "entry" {
		return e, err
	}
	return e, nil
}

func main() {
	c := controller{Controller: common.Controller{Service: "get-entry"}}
	if err := envconfig.Process("journal", &c); err != nil {
		logrus.WithError(err).Panicf("failed to load config")
	}
	common.Init(&c.Controller)

	if c.Environment == "sandbox" {
		c.Controller.ServeSandbox("/entry/:entry_id", "GET", c.getEntry)
	} else {
		lambda.Start(c.getEntry)
	}
}
