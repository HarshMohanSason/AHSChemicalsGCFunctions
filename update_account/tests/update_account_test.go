package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	function "github.com/HarshMohanSason/AHSChemicalsGCFunctions"
)

func TestUpdateAccount(t *testing.T){
	tests := []struct {
		name       string
		input      map[string]interface{}
		method     string
		wantStatus int
	}{
		{
			name: "Empty brands",
			input: map[string]interface{}{
				"uid":        "vUbdNbyWp1OPOW9ITfumWdl8DWe2",
				"brands":     []string{},
				"properties": []string{"2040 N preisker park", "asdb 230 asdfas df", "asdfasdfas"},
			},
			method:     http.MethodPut,
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "Empty properties",
			input: map[string]interface{}{
				"uid":        "vUbdNbyWp1OPOW9ITfumWdl8DWe2",
				"brands":     []string{"pro blend", "macro tech"},
				"properties": []string{},
			},
			method:     http.MethodPut,
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid method (GET instead of PUT)",
			input: map[string]interface{}{
				"uid":        "vUbdNbyWp1OPOW9ITfumWdl8DWe2",
				"brands":     []string{"pro blend", "macro tech"},
				"properties": []string{"2040 N preisker park", "asdb 230 asdfas df", "asdfasdfas"},
			},
			method:     http.MethodGet,
			wantStatus: http.StatusMethodNotAllowed,
		},
		{
			name: "Valid update",
			input: map[string]interface{}{
				"uid":        "vUbdNbyWp1OPOW9ITfumWdl8DWe2",
				"brands":     []string{"new brand", "updated brand"},
				"properties": []string{"new property", "updated property"},
			},
			method:     http.MethodPut,
			wantStatus: http.StatusOK,
		},
		{
			name: "Missing UID field",
			input: map[string]interface{}{
				"brands":     []string{"pro blend"},
				"properties": []string{"2040 N preisker park"},
			},
			method:     http.MethodPut,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests{
		t.Run(tc.name, func(t *testing.T) {
			jsonData, err := json.Marshal(tc.input)
            if err != nil {
                t.Fatalf("Failed to marshal input: %v", err)
            }

            req := httptest.NewRequest(tc.method, "/update-account", bytes.NewReader(jsonData))
            res := httptest.NewRecorder()

           	handler := http.HandlerFunc(function.UpdateAccount)
           	handler.ServeHTTP(res, req)

			if status := res.Code; status != tc.wantStatus {
				t.Errorf("Test %q failed: expected status %v, got %v", tc.name, tc.wantStatus, status)
			}
		})
	}
}