package main

// Please note this is very basic and just meant to be illustrative.
// We strongly recommend that to use anything like this in production, proper error handling, packaging, etc., is required.

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"encoding/json"

	"golang.org/x/oauth2/clientcredentials"
)

// FILL THESE IN
const OAUTH_CLIENT_ID = ""
const OAUTH_CLIENT_SECRET = ""
const LABBIT_DOMAIN = "demo.example.labbit.com"
const OAUTH_DOMAIN = "example-labbit.us.auth0.com"
const OAUTH_AUDIENCE = "urn:labbit.com:example:main_api"

type GetByLabelRequest struct {
	Values []string `json:"values"`
}

type SearchRequest struct {
	Fields map[string][]FieldSearchOperator `json:"searchTermByFields"`
}

type FieldSearchOperator struct {
	Operator string `json:"operator"`
	Value    string `json:"value"`
}

type MultipleEntitiesResponse struct {
	Entities []Entity `json:"entities"`
}

type Entity struct {
	EntityType     string           `json:"type"`
	Fields         map[string]Field `json:"fields"`
	Iri            string           `json:"iri"`
	LabelFieldName string           `json:"labelFieldName"`
	CreatedAtTime  time.Time        `json:"createdAtTime"`
	// this struct is incomplete
}

func (e *Entity) Label() string {
	return e.Fields[e.LabelFieldName].Value
}

type Field struct {
	Value    string `json:"value"`
	DataType string `json:"dataType"`
}

func main() {
	container_name := "LP9181250-QNT"

	conf := &clientcredentials.Config{
		ClientID:       OAUTH_CLIENT_ID,
		ClientSecret:   OAUTH_CLIENT_SECRET,
		TokenURL:       "https://" + OAUTH_DOMAIN + "/oauth/token",
		EndpointParams: url.Values{"audience": {OAUTH_AUDIENCE}},
	}

	ctx := context.Background()
	client := conf.Client(ctx)

	getRequest := GetByLabelRequest{Values: []string{container_name}}

	var getResponse MultipleEntitiesResponse
	api_exchange(client, "https://"+LABBIT_DOMAIN+"/entity/getByLabel", getRequest, &getResponse)

	fmt.Println("Got", len(getResponse.Entities), "entities matching", container_name)

	if len(getResponse.Entities) < 1 {
		os.Exit(2)
	}

	searchRequest := SearchRequest{Fields: map[string][]FieldSearchOperator{"location": {FieldSearchOperator{
		Operator: "==",
		Value:    getResponse.Entities[0].Iri,
	}}}}

	var searchResponse MultipleEntitiesResponse
	api_exchange(client, "https://"+LABBIT_DOMAIN+"/entity/search", searchRequest, &searchResponse)

	for i := 0; i < len(searchResponse.Entities); i++ {
		fmt.Println(searchResponse.Entities[i].Fields["name"].Value)
		fmt.Println(searchResponse.Entities[i].Fields["sublocation"].Value)
		fmt.Println(searchResponse.Entities[i].CreatedAtTime)
		fmt.Println()
	}
}

func api_exchange(client *http.Client, url string, requestObject any, responseObject any) {
	encodedRequest, _ := json.Marshal(requestObject)
	request, _ := http.NewRequest("POST", url, bytes.NewBuffer(encodedRequest))
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")

	fmt.Println("request: ", request.Method, request.URL, request.ContentLength, "bytes")
	response, _ := client.Do(request)

	defer response.Body.Close()
	responseBytes, _ := io.ReadAll(response.Body)

	if response.StatusCode != 200 {
		fmt.Println("Got response", response.StatusCode)
		fmt.Println(string(responseBytes))
		os.Exit(1)
	}

	json.Unmarshal(responseBytes, responseObject)
}
