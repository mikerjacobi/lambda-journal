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

func (c controller) getJournal(ctx context.Context, req *events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	lf := logrus.Fields{"handler": "get_journal"}

	journal, err := c.validate(ctx, req)
	if err != nil {
		logrus.WithError(err).Error("failed to validate journal")
		return common.Response(ctx, http.StatusBadRequest, nil), nil
	}

	item, err := c.DynamoDB.GetItem(journal.GetItem())
	if err != nil {
		logrus.WithError(err).Error("failed to get item")
		return common.Response(ctx, http.StatusNotFound, nil), nil
	}
	journal = journal.GotItem(item)

	logrus.WithFields(lf).Infof("successfully got journal")
	return common.Response(ctx, http.StatusOK, journal), nil
}

func (c controller) validate(ctx context.Context, req *events.APIGatewayProxyRequest) (journal.Journal, error) {
	e := journal.Journal{JournalID: req.PathParameters["journal_id"]}
	res, err := common.GetResource(e.JournalID)
	if err != nil || res != "journal" {
		return e, err
	}
	return e, nil
}

func main() {
	c := controller{Controller: common.Controller{Service: "get-journal"}}
	if err := envconfig.Process("journal", &c); err != nil {
		logrus.WithError(err).Panicf("failed to load config")
	}
	common.Init(&c.Controller)

	if c.Environment == "sandbox" {
		c.Controller.ServeSandbox("/journal/:journal_id", "GET", c.getJournal)
	} else {
		lambda.Start(c.getJournal)
	}
}
