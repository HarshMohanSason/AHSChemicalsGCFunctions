package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"log"
	function "github.com/HarshMohanSason/AHSChemicalsGCFunctions"
)

func TestCreateAccount(t *testing.T){

	userTempData := map[string]interface{}{
		"firstName": "TestFirstName",
		"lastName": "TestLastName",
		"properties": []string{"2040 N preisker lane"}, 
		"brands": []string{"Pro blend"}, 
		"email": "asdsd@gmail.com",
		"password": "testPass",
	}

	//Convert the map to array of bytes
	jsonData, err := json.Marshal(userTempData)
	
	if err != nil{
		t.Errorf("Error encoding the data: %v", err)
	}

	req := httptest.NewRequest("POST", "/create-account", bytes.NewReader(jsonData))

	response := httptest.NewRecorder()

	handler := http.HandlerFunc(function.CreateAccount)
	handler.ServeHTTP(response, req)

	// Check the status code
	body := response.Body.String()
	log.Printf("Response Body: %s", body)

	if status := response.Code; status != http.StatusOK {
		t.Errorf("expected status %v, got %v", http.StatusOK, status)
	}

}