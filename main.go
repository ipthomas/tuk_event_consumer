package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/ipthomas/tukcnst"
	"github.com/ipthomas/tukdbint"
	"github.com/ipthomas/tukutil"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

var initSrvcs = false

func main() {
	lambda.Start(Handle_Request)
}
func Handle_Request(req events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	log.SetFlags(log.Lshortfile)

	var err error
	var dbconn tukdbint.TukDBConnection
	if !initSrvcs {
		dbconn = tukdbint.TukDBConnection{DBUser: os.Getenv(tukcnst.ENV_DB_USER), DBPassword: os.Getenv(tukcnst.ENV_DB_PASSWORD), DBHost: os.Getenv(tukcnst.ENV_DB_HOST), DBPort: os.Getenv(tukcnst.ENV_DB_PORT), DBName: os.Getenv(tukcnst.ENV_DB_NAME)}
		if err = tukdbint.NewDBEvent(&dbconn); err != nil {
			return queryResponse(http.StatusInternalServerError, err.Error(), tukcnst.TEXT_PLAIN)
		}
		initSrvcs = true
	}
	log.Printf("Processing API Gateway %s Request Path %s", req.HTTPMethod, req.Path)

	events := tukdbint.Events{Action: tukcnst.SELECT}
	event := tukdbint.Event{}
	for key, value := range req.QueryStringParameters {
		log.Printf("    %s: %s\n", key, value)
		switch key {
		case tukcnst.ACT:
			events.Action = value
		case tukcnst.QUERY_PARAM_PATHWAY:
			event.Pathway = value
		case tukcnst.QUERY_PARAM_TOPIC:
			event.Topic = value
		case tukcnst.QUERY_PARAM_EXPRESSION:
			event.Expression = value
		case tukcnst.TUK_EVENT_QUERY_PARAM_NOTES:
			event.Comments = value
		case tukcnst.TUK_EVENT_QUERY_PARAM_USER:
			event.User = value
		case tukcnst.TUK_EVENT_QUERY_PARAM_ORG:
			event.Org = value
		case tukcnst.TUK_EVENT_QUERY_PARAM_ROLE:
			event.Role = value
		case tukcnst.TUK_EVENT_QUERY_PARAM_NHS:
			event.NhsId = value
		case tukcnst.TUK_EVENT_QUERY_PARAM_VERSION:
			event.Version = tukutil.GetIntFromString(value)
		case tukcnst.TUK_EVENT_QUERY_PARAM_TASK_ID:
			event.TaskId = tukutil.GetIntFromString(value)
		}
	}
	if req.Body != "" {
		event.Comments = req.Body
	}
	if events.Action == tukcnst.INSERT {
		event.Authors = event.User + " " + event.Org + " " + event.Role
		event.Speciality = event.Role
	}
	events.Events = append(events.Events, event)
	tukdbint.NewDBEvent(&events)
	rsp, _ := json.MarshalIndent(events, "", "  ")
	return queryResponse(http.StatusOK, string(rsp), tukcnst.APPLICATION_JSON)

}
func setAwsResponseHeaders(contentType string) map[string]string {
	awsHeaders := make(map[string]string)
	awsHeaders["Server"] = "TUK_Event_Consumer_Proxy"
	awsHeaders["Access-Control-Allow-Origin"] = "*"
	awsHeaders["Access-Control-Allow-Headers"] = "accept, Content-Type"
	awsHeaders["Access-Control-Allow-Methods"] = "GET, POST, OPTIONS"
	awsHeaders[tukcnst.CONTENT_TYPE] = contentType
	return awsHeaders
}
func queryResponse(statusCode int, body string, contentType string) (*events.APIGatewayProxyResponse, error) {
	log.Println(body)
	return &events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers:    setAwsResponseHeaders(contentType),
		Body:       body,
	}, nil
}
