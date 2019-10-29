package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/sirupsen/logrus"
)

type APIResponse struct {
	Code    int    `json:"code"`
	Type    string `json:"type"`
	Message string `json:"message"`
}

func JSONApiResponse(w http.ResponseWriter, message string, statusCode int) {
	apiResp := APIResponse{
		Code:    statusCode,
		Type:    http.StatusText(statusCode),
		Message: message,
	}
	output, err := json.Marshal(apiResp)
	if err != nil {
		logrus.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	_, err = w.Write(output)
	if err != nil {
		logrus.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func JSONResponse(w http.ResponseWriter, output []byte, statusCode int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	_, err := w.Write(output)
	if err != nil {
		logrus.Error(err)
		JSONApiResponse(w, err.Error(), http.StatusInternalServerError)
	}
}
