package main

import (
	"go.uber.org/zap"
	"io"
	"net/http"
)

func GrpcClient(log *zap.Logger) {

	resp, err := http.Get("http://hello.default.svc.cluster.local")
	if err != nil {
		log.Error("Failed to connect to hello.default.svc.cluster.local", zap.Error(err))
	}

	if resp == nil {
		log.Fatal("response is empty")
	}

	defer func(Body io.ReadCloser) {
		er := Body.Close()
		if er != nil {
			log.Error("error closing the body")
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("Failed to read body", zap.Error(err))
	}

	if body != nil {
		log.Info("body", zap.String("body", string(body)))
	}
}
