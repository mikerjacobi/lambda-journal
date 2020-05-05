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

func (c controller) insertEntry(ctx context.Context, req *events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	lf := logrus.Fields{"handler": "insert_entry"}

	entry, err := c.validate(ctx, req)
	if err != nil {
		logrus.WithError(err).Error("failed to validate entry")
		return common.Response(ctx, http.StatusBadRequest, nil), nil
	}

	if _, err := c.DynamoDB.PutItem(entry.PutItem()); err != nil {
		logrus.WithError(err).Error("failed to put item")
		return common.Response(ctx, http.StatusOK, nil), nil
	}

	logrus.WithFields(lf).Infof("successfully inserted entry")
	return common.Response(ctx, http.StatusOK, entry), nil
}

func (c controller) validate(ctx context.Context, req *events.APIGatewayProxyRequest) (journal.Entry, error) {
	e := journal.Entry{
		EntryID: "entry_" + uuid.New().String(),
	}
	if err := json.Unmarshal([]byte(req.Body), &e); err != nil {
		return e, errors.Wrap(err, "failed to unmarshal")
	}

	errs := []string{}
	if e.Entry == "" {
		errs = append(errs, fmt.Sprintf("entry cannot be empty"))
	}
	if len(errs) > 0 {
		return e, errors.New(strings.Join(errs, ";"))
	}
	e.Created = time.Now().Format(time.RFC3339)
	e.Updated = time.Now().Format(time.RFC3339)
	return e, nil
}

func main() {
	c := controller{Controller: common.Controller{Service: "insert_entry"}}
	if err := envconfig.Process("journal", &c); err != nil {
		logrus.WithError(err).Panicf("failed to load config")
	}
	common.Init(&c.Controller)

	if c.Environment == "sandbox" {
		c.Controller.ServeSandbox("/entry", "POST", c.insertEntry)
	} else {
		lambda.Start(c.insertEntry)
	}
}
