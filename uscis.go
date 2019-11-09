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

func getUSCIS(ticket string) (string, error) {
	body := fmt.Sprintf("hangeLocale=&appReceiptNum=%s&initCaseSearch=CHECK+STATUS", ticket)
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
		fmt.Println(getUSCIS(os.Args[1]))
	}

}
