package main

import (
	"log"
	"net/http"
	"os"

	function "github.com/HarshMohanSason/AHSChemicalsGCFunctions"
	"github.com/HarshMohanSason/AHSChemicalsGCShared/shared"
	"github.com/joho/godotenv"
)
func main(){

	//Only for local development
	if os.Getenv("ENV") == "DEBUG"{
		
		//Load the env file
		err := godotenv.Load("../keys/.env")
		if err != nil{
			log.Printf("Error occurred loading the env file: %v", err)
		}

		adminSDKFilePath := os.Getenv("FIREBASE_CREDENTIALS_DEBUG")

		//Initialize the debug project sdk
		shared.InitFirebaseDebug(adminSDKFilePath)
		
		http.Handle("/fetch-accounts", http.HandlerFunc(function.FetchAccounts))
			
		log.Print("fetch-accounts started at: 3004")
		err = http.ListenAndServe(":3004", nil)
		if err != nil{
			log.Printf("Error occurred when starting the server: %v", err)
		} 
	}
}