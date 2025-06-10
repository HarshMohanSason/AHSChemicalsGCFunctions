package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	function "github.com/HarshMohanSason/AHSChemicalsGCFunctions"
)


func TestSendMail(t *testing.T){

	testCase := map[string]any{
		"recipients": map[string]string{
			"test@123gmail.com": "Harsh",
		},
		"template_id": os.Getenv("SENDGRID_TEST_TEMPLATE_ID"), 
		"data": map[string]any{
			"name": "Harsh Mohan Sason", 
			"email": "abcd@gmail.com", 
			"password": "TestPass",
		},
	}

	body, err := json.Marshal(testCase)
	if err != nil {
		t.Fatalf("Error occurred encoding the body: %v", err)
	}

	req := httptest.NewRequest("POST", "/send-mail", bytes.NewReader(body))
	response := httptest.NewRecorder()

	handler := http.HandlerFunc(function.SendMail) // replace with actual handler name
	handler.ServeHTTP(response, req)

	// Check the status code
	if status := response.Code; status != http.StatusOK {
		t.Errorf("expected status %v, got %v", http.StatusOK, status)
	}
}