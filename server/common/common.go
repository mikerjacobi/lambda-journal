package common

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/husobee/vestigo"
	"github.com/sirupsen/logrus"
)

func init() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
}

type APIGatewayHandler = func(context.Context, *events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error)

//Controller is the common controller type that lambda handlers inherit
type Controller struct {
	Service string
	Configuration
	Handler APIGatewayHandler
	*dynamodb.DynamoDB
}

type Configuration struct {
	Environment string `envconfig:"JOURNAL_ENVIRONMENT"`
	BaseDomain  string `envconfig:"JOURNAL_BASE_DOMAIN"` //"http://dev.jaqobi.com"
}

func Init(c *Controller) {
	awsConfig := &aws.Config{Region: aws.String("us-west-2")}
	if c.Environment == "sandbox" {
		awsConfig.Endpoint = aws.String(c.BaseDomain + ":8000")
	}

	sess := session.Must(session.NewSession(awsConfig))
	c.DynamoDB = dynamodb.New(sess)
}

func (c Controller) ServeSandbox(route, method string, handler APIGatewayHandler) {
	c.Handler = handler
	router := vestigo.NewRouter()
	router.SetGlobalCors(&vestigo.CorsAccessControl{
		AllowOrigin:  []string{"*"},
		AllowHeaders: []string{"content-type", "cognitoauthenticationprovider", "cognitoidentityid"},
	})
	router.Add(method, route, c.SandboxHandler, []vestigo.Middleware{}...)

	// /certs is volume mounted in sandbox.yaml, and the .pems came from letsencrypt
	//cert := "/certs/fullchain1.pem"
	//key := "/certs/privkey1.pem"

	logrus.Infof("serving sandbox route (%s). method:(%s) route:(%s) port:(80)", c.Service, method, route)
	//logrus.Fatal(http.ListenAndServeTLS("0.0.0.0:443",	 cert,	 key,	router))
	logrus.Fatal(http.ListenAndServe("0.0.0.0:80", router))
}

func (c Controller) SandboxHandler(rw http.ResponseWriter, r *http.Request) {
	//setup get parameters
	params := map[string]string{}
	for k := range r.URL.Query() {
		params[strings.Replace(k, ":", "", -1)] = r.URL.Query()[k][0]
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		logrus.WithError(err).Errorf("failed to read body")
		return
	}
	req := &events.APIGatewayProxyRequest{
		PathParameters:        params,
		QueryStringParameters: params,
		Body:                  string(body),
		/*
			RequestContext: events.APIGatewayProxyRequestContext{
				Identity: events.APIGatewayRequestIdentity{
					CognitoIdentityID:             r.Header.Get("cognitoIdentityId"),
					CognitoAuthenticationProvider: r.Header.Get("cognitoAuthenticationProvider"),
				},
			},
		*/
	}

	resp, err := c.Handler(context.Background(), req)
	rw.WriteHeader(resp.StatusCode) //resp should never be empty
	if err != nil {
		logrus.WithError(err).Errorf("failed to execute handler")
		return
	}
	rw.Write([]byte(resp.Body))
}

//Response constructor for api gateway response
func Response(ctx context.Context, code int, responsePayload interface{}) *events.APIGatewayProxyResponse {
	contentType := "application/json"
	if ct, ok := ctx.Value(ctxContentType).(string); ok {
		contentType = ct
	}
	headers := map[string]string{
		"Content-Type":                     contentType,
		"Access-Control-Allow-Origin":      "*",
		"Access-Control-Allow-Credentials": "true",
	}
	body, err := json.Marshal(responsePayload)
	if err != nil {
		logrus.WithError(err).Error("failed to marshal response payload")
		return &events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Headers:    headers,
			Body:       `{"error": "internal error"}`,
		}
	}

	return &events.APIGatewayProxyResponse{StatusCode: code, Headers: headers, Body: string(body)}
}

func GetResource(resourceID string) (string, error) {
	split := strings.Split(resourceID, "_")
	if len(split) != 2 {
		return "", fmt.Errorf("invalid resource format: %s", resourceID)
	}
	resourceType, id := split[0], split[1]
	if !govalidator.IsUUID(id) {
		return "", fmt.Errorf("invalid resource id: %s", resourceID)
	}
	return resourceType, nil
}
