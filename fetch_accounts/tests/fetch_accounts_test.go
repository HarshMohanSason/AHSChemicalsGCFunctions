package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	function "github.com/HarshMohanSason/AHSChemicalsGCFunctions"
)

func TestFetchAccounts(t *testing.T){

	req := httptest.NewRequest(http.MethodGet, "/fetch-accounts", nil)

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(function.FetchAccounts) // replace with actual handler name
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("expected status %v, got %v", http.StatusOK, status)
	}

}