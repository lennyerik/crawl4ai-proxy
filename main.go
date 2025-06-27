package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

const IP = "" // leave blank for all addresses
const PORT = 8000
const CRAWL4AIENDPOINT = "http://crawl4ai:11235/crawl"

// const CRAWL4AIENDPOINT = "http://localhost:11235/crawl"

// For the openwebui-facing endpoint
type Request struct {
	Urls []string `json:"urls"`
}

type SuccessResponseItem struct {
	PageContent string            `json:"page_content"`
	Metadata    map[string]string `json:"metadata"`
}
type SuccessResponse []SuccessResponseItem

type ErrorResponse struct {
	ErrorName string `json:"error"`
	Detail    string `json:"detail"`
}

// For the crawl4ai-facing endpoint
type CrawlResponse struct {
	Results []struct {
		Url      string `json:"url"`
		Markdown struct {
			RawMarkdown string `json:"raw_markdown"`
		} `json:"markdown"`
		Metadata map[string]string `json:"metadata"`
	} `json:"results"`
}

func errorResponseFromError(name string, err error) ErrorResponse {
	return ErrorResponse{
		ErrorName: name,
		Detail:    err.Error(),
	}
}

func jsonEncodeInfallible(object any) []byte {
	encoded, err := json.Marshal(object)
	if err != nil {
		panic(err)
	}
	return encoded
}

func CrawlEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Content-Type", "application/json")

	if request.Method != "POST" {
		response.WriteHeader(405)
		resp := ErrorResponse{ErrorName: "method not allowed"}
		response.Write(jsonEncodeInfallible(resp))
		log.Printf("405 method not allowed :: %s\n", request.RemoteAddr)
		return
	}

	if request.Header.Get("Content-Type") != "application/json" {
		response.WriteHeader(400)
		resp := ErrorResponse{ErrorName: "content type must be application/json"}
		response.Write(jsonEncodeInfallible(resp))
		log.Printf("400 invalid content type :: %s\n", request.RemoteAddr)
		return
	}

	var requestData Request
	err := json.NewDecoder(request.Body).Decode(&requestData)
	if err != nil {
		response.WriteHeader(400)
		resp := errorResponseFromError("invalid json", err)
		response.Write(jsonEncodeInfallible(resp))
		log.Printf("400 invalid json :: %s\n", request.RemoteAddr)
		return
	}

	log.Printf("Request to crawl %s from %s\n", requestData.Urls, request.RemoteAddr)

	req, err := http.NewRequest("POST", CRAWL4AIENDPOINT, bytes.NewReader(jsonEncodeInfallible(requestData)))
	if err != nil {
		panic(err)
	}

	crawlResponse, err := http.DefaultClient.Do(req)
	if err != nil || crawlResponse.StatusCode != 200 {
		response.WriteHeader(502)
		resp := ErrorResponse{ErrorName: "bad gateway"}
		response.Write(jsonEncodeInfallible(resp))
		log.Printf("502 bad gateway :: %s\n", request.RemoteAddr)
		return
	}

	var crawlData CrawlResponse
	err = json.NewDecoder(crawlResponse.Body).Decode(&crawlData)
	if err != nil {
		response.WriteHeader(502)
		resp := ErrorResponse{ErrorName: "bad gateway", Detail: "invalid json received from crawl api"}
		response.Write(jsonEncodeInfallible(resp))
		log.Printf("502 bad gateway - invalid json from crawl api :: %s\n", request.RemoteAddr)
		return
	}

	ret := SuccessResponse{}
	for _, result := range crawlData.Results {
		if result.Metadata == nil {
			result.Metadata = map[string]string{}
		}

		for key, value := range result.Metadata {
			if value == "" {
				delete(result.Metadata, key)
			}
		}

		result.Metadata["source"] = result.Url

		ret = append(ret, SuccessResponseItem{
			PageContent: result.Markdown.RawMarkdown,
			Metadata:    result.Metadata,
		})
	}

	response.WriteHeader(200)
	response.Write(jsonEncodeInfallible(ret))
	log.Printf("200 :: %s\n", request.RemoteAddr)
}

func main() {
	http.HandleFunc("/crawl", CrawlEndpoint)

	listenAddress := fmt.Sprintf("%s:%d", IP, PORT)
	log.Printf("Listening on %s\n", listenAddress)

	err := http.ListenAndServe(listenAddress, nil)
	if err != nil {
		log.Println(err)
	}
}
