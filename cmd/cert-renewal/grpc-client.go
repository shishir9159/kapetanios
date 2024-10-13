package main

import (
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"io"
	"net/http"
)

type ResponseDTO struct {
	Message string `json:"message"`
}

func GrpcClient(log *zap.Logger) {

	client := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, "http://hello", nil)
	if req != nil {
		req.Header.Add("Content-Type", "application/json")
	} else {
		log.Error(fmt.Sprintf("error in requesting: %s", err.Error()))
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Error(fmt.Sprintf("error in getting the response body: %s", err.Error()))
	}

	defer resp.Body.Close()
	if resp.Body != nil {
		jsonDataFromHttp, er := io.ReadAll(resp.Body)
		if er != nil {
			log.Error(fmt.Sprintf("error in reading response body: %s", er.Error()))
		}
		log.Info("response", zap.String("jsonDataFromHttp", string(jsonDataFromHttp)))
		responseDto := ResponseDTO{}
		er = json.Unmarshal(jsonDataFromHttp, &responseDto)
		if er != nil {
			log.Error(fmt.Sprintf("error in parsing response body to json: %s", er.Error()))
		}
		log.Info("responseDTO", zap.String("r", responseDto.Message))
	}

	resp, err = http.Get("http://hello.default")

	if req != nil {
		req.Header.Add("Content-Type", "application/json")
	}
	if err != nil {
		log.Error("Failed to connect to hello.default.svc.cluster.local", zap.Error(err))
	}

	if resp == nil {
		log.Fatal("response is empty")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("Failed to read body", zap.Error(err))
	}

	log.Info("body", zap.String("body", string(body)))
}
