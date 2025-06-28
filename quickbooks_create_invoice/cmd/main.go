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
		//Register firebase 
		shared.InitFirebaseDebug(os.Getenv("FIREBASE_CREDENTIALS_DEBUG"))
		
		//Register quickbooks credentials
		shared.InitQuickBooksDebug()

		http.Handle("/quickbooks-create-invoice", http.HandlerFunc(function.CreateInvoice))
			
		log.Print("quickbooks-create-invoice started at: 4002")
		err = http.ListenAndServe(":4002", nil)
		if err != nil{
			log.Printf("Error occurred when starting the server: %v", err)
		} 
	}
}