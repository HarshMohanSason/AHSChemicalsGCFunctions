package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	function "github.com/HarshMohanSason/AHSChemicalsGCFunctions"
)

func TestDeleteAccount(t *testing.T) {
    tests := []struct {
        name       string
        input      map[string]interface{}
        method     string
        wantStatus int
        wantBody   string 
    }{
        {
            name: "Valid request",
            input: map[string]interface{}{
                "uid": "AValidUIDGoesHere",
            },
            method:     http.MethodDelete,
            wantStatus: http.StatusOK,
        },
        {
            name: "Missing UID",
            input:      map[string]interface{}{},
            method:     http.MethodDelete,
            wantStatus: http.StatusBadRequest,
        },
        {
            name: "Wrong HTTP method",
            input: map[string]interface{}{
                "uid": "0Wy5UXyUojQHbt8FnwJycjBiKPZ2",
            },
            method:     http.MethodGet,
            wantStatus: http.StatusMethodNotAllowed,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            jsonData, err := json.Marshal(tt.input)
            if err != nil {
                t.Fatalf("Failed to marshal input: %v", err)
            }

            req := httptest.NewRequest(tt.method, "/delete-account", bytes.NewReader(jsonData))
            res := httptest.NewRecorder()

            handler := http.HandlerFunc(function.DeleteAccount)
            handler.ServeHTTP(res, req)

            if res.Code != tt.wantStatus {
                t.Errorf("[%s] Expected status %v, got %v", tt.name, tt.wantStatus, res.Code)
            }
        })
    }
}