package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/google/uuid"
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

func (c controller) insertJournal(ctx context.Context, req *events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	lf := logrus.Fields{"handler": "insert_journal"}

	journal, err := c.validate(ctx, req)
	if err != nil {
		logrus.WithError(err).Error("failed to validate journal")
		return common.Response(ctx, http.StatusBadRequest, nil), nil
	}

	if _, err := c.DynamoDB.PutItem(journal.PutItem()); err != nil {
		logrus.WithError(err).Error("failed to put item")
		return common.Response(ctx, http.StatusOK, nil), nil
	}

	logrus.WithFields(lf).Infof("successfully inserted journal")
	return common.Response(ctx, http.StatusOK, journal), nil
}

func (c controller) validate(ctx context.Context, req *events.APIGatewayProxyRequest) (journal.Journal, error) {
	e := journal.Journal{JournalID: "journal:" + uuid.New().String()}
	if err := json.Unmarshal([]byte(req.Body), &e); err != nil {
		return e, errors.Wrap(err, "failed to unmarshal")
	}

	errs := []string{}
	if e.Entry == "" {
		errs = append(errs, fmt.Sprintf("entry cannot be empty"))
	}
	if e.Created == "" {
		errs = append(errs, fmt.Sprintf("created cannot be empty"))
	}
	if len(errs) > 0 {
		return e, errors.New(strings.Join(errs, ";"))
	}
	e.Updated = time.Now().Format(time.RFC3339)
	return e, nil
}

func main() {
	c := controller{Controller: common.Controller{Service: "insert_journal"}}
	if err := envconfig.Process("journal", &c); err != nil {
		logrus.WithError(err).Panicf("failed to load config")
	}
	common.Init(&c.Controller)

	if c.Environment == "sandbox" {
		c.Controller.ServeSandbox("/journal", "POST", c.insertJournal)
	} else {
		lambda.Start(c.insertJournal)
	}
}
