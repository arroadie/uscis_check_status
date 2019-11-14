package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/jbowtie/gokogiri"
)

// Event the aws event that contains the request
type Event struct {
	Ticket      string `json:"ticket"`
	Message     string `json:"message"`
	Destination string `json:"destination"`
}

func getUSCIS(event Event) (Event, error) {
	body := fmt.Sprintf("hangeLocale=&appReceiptNum=%s&initCaseSearch=CHECK+STATUS", event.Ticket)
	req, err := http.NewRequest("POST", "https://egov.uscis.gov/casestatus/mycasestatus.do", bytes.NewBuffer([]byte(body)))
	if err != nil {
		return event, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return event, err
	}
	defer resp.Body.Close()

	page, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return event, err
	}

	doc, err := gokogiri.ParseHtml(page)
	if err != nil {
		return event, err
	}

	html := doc.Root().FirstChild()
	defer doc.Free()
	results, _ := html.Search("/html/body/div[2]/form/div/div[1]/div/div/div[2]/div[3]/p")
	event.Message = results[0].Content()
	return event, nil
}

func lambdaWrapper(event Event) (Event, error) {
	// get info in dynamo by ticket
	// get info from uscis
	// check if message is the same
	// if different
	//   update the event message
	//   store the new event message
	//   notify with the new event message
	// else
	//   return

	sess, err := session.NewSession()
	if err != nil {
		panic(err)
	}

	svc := dynamodb.New(sess)

	getTicket := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"ticket": {
				S: aws.String(event.Ticket),
			},
		},
		TableName: aws.String("tickets"),
	}

	dbResponse, err := svc.GetItem(getTicket)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case dynamodb.ErrCodeResourceNotFoundException:
				log.Println(dynamodb.ErrCodeResourceNotFoundException, aerr.Error())
			default:
				// TODO: Find why the validation exception falls here...
				log.Println(aerr.Error())
			}
		} else {
			log.Println(err)
		}
	}

	event, err = getUSCIS(event)
	if err != nil {
		log.Println(err)
		panic(err)
	}

	saveToDynamo(event, svc)

	if dbResponse.Item["message"] == nil || event.Message != *dbResponse.Item["message"].S {

		snsClient := sns.New(sess)
		snsInput := &sns.PublishInput{
			Message:     aws.String(event.Message),
			PhoneNumber: aws.String(os.Getenv("DESTINATION_NUMBER")),
		}
		_, err = snsClient.Publish(snsInput)
		if err != nil {
			log.Println("Publish error:", err)
		}
	}

	return event, nil
}

func saveToDynamo(event Event, svc *dynamodb.DynamoDB) {
	av, err := dynamodbattribute.MarshalMap(event)
	item := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String("tickets"),
	}

	_, err = svc.PutItem(item)
	if err != nil {
		log.Println("Error calling PutItem:")
		log.Println(err.Error())
	}
}

func main() {

	if len(os.Getenv("LAMBDA")) > 0 {
		lambda.Start(lambdaWrapper)
	} else {
		if len(os.Args) > 1 {
			response, err := getUSCIS(Event{os.Args[1], "", ""})
			if err != nil {
				panic(err)
			}
			fmt.Println(response.Message)
		} else {
			fmt.Println("Argument needed to execute the script (case identifier).")
		}
	}

}
