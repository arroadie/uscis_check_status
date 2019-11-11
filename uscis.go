package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/jbowtie/gokogiri"
)

// Event the aws event that contains the request
type Event struct {
	Ticket string `json:"ticket"`
}

func getUSCIS(event Event) (string, error) {
	body := fmt.Sprintf("hangeLocale=&appReceiptNum=%s&initCaseSearch=CHECK+STATUS", event.Ticket)
	req, err := http.NewRequest("POST", "https://egov.uscis.gov/casestatus/mycasestatus.do", bytes.NewBuffer([]byte(body)))
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	page, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	doc, err := gokogiri.ParseHtml(page)
	if err != nil {
		panic(err)
	}

	html := doc.Root().FirstChild()
	defer doc.Free()
	results, _ := html.Search("/html/body/div[2]/form/div/div[1]/div/div/div[2]/div[3]/p")
	return fmt.Sprintln(results[0].Content()), nil
}

func main() {

	if len(os.Getenv("LAMBDA")) > 0 {
		lambda.Start(getUSCIS)
	} else {
		if len(os.Args) > 1 {
			response, err := getUSCIS(Event{os.Args[1]})
			if err != nil {
				panic(err)
			}
			fmt.Println(response)
		} else {
			fmt.Println("Argument needed to execute the script (case identifier).")
		}
	}

}
