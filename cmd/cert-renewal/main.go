package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type ResponseDTO struct {
	Message string `json:"message"`
}

func main() {
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, "http://hello.default.svc.cluster.local", nil)
	if req != nil {
		req.Header.Add("Content-Type", "application/json")
	} else {
		log.Print(fmt.Sprintf("error in requesting: %s", err.Error()))
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Print(fmt.Sprintf("error in getting the response body: %s", err.Error()))
	}

	defer func(Body io.ReadCloser) {
		er := Body.Close()
		if er != nil {
			log.Print(fmt.Sprintf("error closing response body: %s", er.Error()))
		}
	}(resp.Body)

	if resp.Body != nil {
		jsonDataFromHttp, er := io.ReadAll(resp.Body)
		if er != nil {
			log.Print(fmt.Sprintf("error in reading response body: %s", er.Error()))
		}
		log.Print("response")
		fmt.Println(jsonDataFromHttp)
		responseDto := ResponseDTO{}
		er = json.Unmarshal(jsonDataFromHttp, &responseDto)
		if er != nil {
			log.Print(fmt.Sprintf("error in parsing response body to json: %s", er.Error()))
		}
		log.Print("responseDTO")
		log.Println(responseDto.Message)
	}
}
