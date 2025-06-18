package tests

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	function "github.com/HarshMohanSason/AHSChemicalsGCFunctions"
)

func TestCreateAccount(t *testing.T) {
	testCases := []struct {
		Name           string
		Input          function.CreateAccountRequest
		ExpectedStatus int
	}{
		{
			Name: "Valid Request",
			Input: function.CreateAccountRequest{
				Name:        "Test User",
				PhoneNumber: "+11231231211",
				Email:       "testuser@example.com",
				Password:    "ValidPass123",
				Properties: []map[string]string{
					{
						"city":   "San Jose",
						"county": "Santa Clara",
						"state":  "California",
						"postal": "95112",
						"street": "123 Main St",
					},
				},
				Brands: []string{"ProBlend"},
			},
			ExpectedStatus: http.StatusOK,
		},
		{
			Name: "Missing Properties Array",
			Input: function.CreateAccountRequest{
				Name:        "Test User",
				PhoneNumber: "+15595484965",
				Email:       "testuser@example.com",
				Password:    "ValidPass123",
				Properties:  []map[string]string{},
				Brands:      []string{"ProBlend"},
			},
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			Name: "Missing Brand Array",
			Input: function.CreateAccountRequest{
				Name:        "Test User",
				PhoneNumber: "+15595484965",
				Email:       "testuser@example.com",
				Password:    "ValidPass123",
				Properties: []map[string]string{
					{
						"city":   "San Jose",
						"county": "Santa Clara",
						"state":  "California",
						"postal": "95112",
						"street": "123 Main St",
					},
				},
				Brands: []string{},
			},
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			Name: "Missing Required Property Field (Street)",
			Input: function.CreateAccountRequest{
				Name:        "Test User",
				PhoneNumber: "+15595484965",
				Email:       "testuser@example.com",
				Password:    "ValidPass123",
				Properties: []map[string]string{
					{
						"city":   "San Jose",
						"county": "Santa Clara",
						"state":  "California",
						"postal": "95112",
						// Missing "street"
					},
				},
				Brands: []string{"ProBlend"},
			},
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			Name: "Empty Phone Number",
			Input: function.CreateAccountRequest{
				Name:        "Test User",
				PhoneNumber: "",
				Email:       "testuser@example.com",
				Password:    "ValidPass123",
				Properties: []map[string]string{
					{
						"city":   "San Jose",
						"county": "Santa Clara",
						"state":  "California",
						"postal": "95112",
						"street": "123 Main St",
					},
				},
				Brands: []string{"ProBlend"},
			},
			ExpectedStatus: http.StatusInternalServerError,
		},
		{
			Name: "Invalid Email Format",
			Input: function.CreateAccountRequest{
				Name:        "Test User",
				PhoneNumber: "+15595484965",
				Email:       "invalid-email-format",
				Password:    "ValidPass123",
				Properties: []map[string]string{
					{
						"city":   "San Jose",
						"county": "Santa Clara",
						"state":  "California",
						"postal": "95112",
						"street": "123 Main St",
					},
				},
				Brands: []string{"ProBlend"},
			},
			ExpectedStatus: http.StatusInternalServerError,
		},
		{
			Name: "Weak Password",
			Input: function.CreateAccountRequest{
				Name:        "Test User",
				PhoneNumber: "+15595484965",
				Email:       "testuser@example.com",
				Password:    "123",
				Properties: []map[string]string{
					{
						"city":   "San Jose",
						"county": "Santa Clara",
						"state":  "California",
						"postal": "95112",
						"street": "123 Main St",
					},
				},
				Brands: []string{"ProBlend"},
			},
			ExpectedStatus: http.StatusInternalServerError,
		},
	}

	//Running the testcases 
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			jsonData, err := json.Marshal(testCase.Input)
			if err != nil {
				t.Errorf("Error occurred marshalling the json data %v", err)
			}
			req := httptest.NewRequest("POST", "/create-account", bytes.NewReader(jsonData))
			response := httptest.NewRecorder()

			handler := http.HandlerFunc(function.CreateAccount)
			handler.ServeHTTP(response, req)
			body := response.Body
			if status := response.Code; status != testCase.ExpectedStatus{
				log.Print(body)
				t.Errorf("expected status %v, got %v", testCase.ExpectedStatus, status)
			}
		})
	}
}

