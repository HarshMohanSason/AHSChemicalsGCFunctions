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
		
		http.Handle("/delete-account", http.HandlerFunc(function.DeleteAccount))
			
		log.Print("delete-account started at: 3001")
		err = http.ListenAndServe(":3001", nil)
		if err != nil{
			log.Printf("Error occurred when starting the server: %v", err)
		} 
	}
}