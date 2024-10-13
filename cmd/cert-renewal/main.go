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

	defer resp.Body.Close()
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
		fmt.Println(responseDto.Message)
	}

	resp, err = http.Get("http://hello.default")

	if req != nil {
		req.Header.Add("Content-Type", "application/json")
	}
	if err != nil {
		log.Print("Failed to connect to hello.default.svc.cluster.local", err.Error())
	}

	if resp == nil {
		log.Fatal("response is empty")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Print("Failed to read body", err.Error())
	}

	log.Print("body")
	fmt.Println(body)
}
