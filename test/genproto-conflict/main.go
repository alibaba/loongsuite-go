package main

import (
	"go.uber.org/zap"
	"net/http"
	_ "github.com/aliyun/aliyun-odps-go-sdk/odps"
)

func main() {
	http.HandleFunc("/log", func(w http.ResponseWriter, r *http.Request) {
		logger := zap.NewExample()
		logger.Debug("this is debug message")
		logger.Info("this is info message")
		logger.Warn("this is warn message")
		logger.Error("this is error message")
	})

	http.ListenAndServe(":9999", nil)
}
