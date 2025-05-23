package utils

import (
	"net/http" 
	"os")

func HandleCors(response http.ResponseWriter, request *http.Request) bool {
	
	var origin string
	if os.Getenv("ENV") == "DEV" {
		origin = "http://localhost:3000"
	} else {
		origin = "https://ahschemicalsdebug.web.app"
	}

	// Cors ehaders
	response.Header().Set("Access-Control-Allow-Origin", origin)
	response.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	response.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	response.Header().Set("Access-Control-Allow-Credentials", "true")

	// Pre flight request
	if request.Method == http.MethodOptions {
		response.WriteHeader(http.StatusOK)
		return true
	}

	return false 
}