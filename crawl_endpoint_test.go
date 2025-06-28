package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"testing"
)

func init() {
	log.SetOutput(io.Discard)
}

type TestResponse struct {
	header     http.Header
	StatusCode int
	Content    []byte
}

func (response *TestResponse) Header() http.Header {
	return response.header
}

func (response *TestResponse) Write(content []byte) (int, error) {
	response.Content = content
	return len(content), nil
}

func (response *TestResponse) WriteHeader(statusCode int) {
	response.StatusCode = statusCode
}

func (response *TestResponse) DecodeJson() (map[string]any, error) {
	var data map[string]any
	err := json.Unmarshal(response.Content, &data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func callEndpoint(request *http.Request) *TestResponse {
	response := TestResponse{
		header: http.Header{},
	}
	CrawlEndpoint(&response, request)
	return &response
}

func expectErrorWithName(t *testing.T, response *TestResponse, expectedErrorName string) bool {
	jsonData, err := response.DecodeJson()
	if err != nil {
		t.Error(err)
		return false
	}

	errorName, exists := jsonData["error"]
	if !exists {
		t.Error("error property does not exist on response")
		return false
	}

	errorName, isString := errorName.(string)
	if !isString {
		t.Error("error property is not a string")
		return false
	}

	if errorName != expectedErrorName {
		t.Errorf("Incorrect error received: %s", errorName)
		return false
	}

	return true
}

func TestMethodNotAllowed(t *testing.T) {
	request, err := http.NewRequest("GET", "/crawl", nil)
	if err != nil {
		panic(err)
	}
	response := callEndpoint(request)

	if response.StatusCode != 405 {
		t.Errorf("Got status code %d", response.StatusCode)
		return
	}

	expectErrorWithName(t, response, "method not allowed")
}

func TestInvalidContentType(t *testing.T) {
	request, err := http.NewRequest("POST", "/crawl", strings.NewReader("{}"))
	if err != nil {
		panic(err)
	}
	request.Header.Add("Content-Type", "application/pdf")
	response := callEndpoint(request)

	if response.StatusCode != 400 {
		t.Errorf("Got status code %d", response.StatusCode)
		return
	}

	expectErrorWithName(t, response, "content type must be application/json")
}

func TestInvalidJson(t *testing.T) {
	request, err := http.NewRequest("POST", "/crawl", strings.NewReader("hello, world"))
	if err != nil {
		panic(err)
	}
	request.Header.Add("Content-Type", "application/json")
	response := callEndpoint(request)

	if response.StatusCode != 400 {
		t.Errorf("Got status code %d", response.StatusCode)
		return
	}

	if !expectErrorWithName(t, response, "invalid json") {
		return
	}

	// Send a string instead of an array of urls
	request, err = http.NewRequest("POST", "/crawl", strings.NewReader("{\"urls\": \"hello!\"}"))
	if err != nil {
		panic(err)
	}
	request.Header.Add("Content-Type", "application/json")
	response = callEndpoint(request)

	if response.StatusCode != 400 {
		t.Errorf("Got status code %d", response.StatusCode)
		return
	}

	expectErrorWithName(t, response, "invalid json")
}
