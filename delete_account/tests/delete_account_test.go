package tests

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	function "github.com/HarshMohanSason/AHSChemicalsGCFunctions"
)

func TestDeleteAccount(t *testing.T) {
    tests := []struct {
        name       string
        uid        string
        method     string
        wantStatus int
    }{
        {
            name: "Valid request",
            uid: "kAmndAeJPmYqZ9B7w95UbKNtKsn1",
            method:     http.MethodDelete,
            wantStatus: http.StatusOK,
        },
        {
            name: "Missing UID",
            uid: "",
            method:     http.MethodDelete,
            wantStatus: http.StatusBadRequest,
        },
        {
            name: "Wrong HTTP method",
            uid: "kAmndAeJPmYqZ9B7w95UbKNtKsn1",
            method:     http.MethodGet,
            wantStatus: http.StatusMethodNotAllowed,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            
            reqPath := fmt.Sprintf("/delete-account/?uid=%s", tt.uid)
            req := httptest.NewRequest(tt.method, reqPath, nil)
            res := httptest.NewRecorder()

            handler := http.HandlerFunc(function.DeleteAccount)
            handler.ServeHTTP(res, req)

            if res.Code != tt.wantStatus {
                t.Errorf("[%s] Expected status %v, got %v", tt.name, tt.wantStatus, res.Code)
            }
        })
    }
}