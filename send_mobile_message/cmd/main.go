package main

import (
	"log"
	"net/http"
	"os"

	function "github.com/HarshMohanSason/AHSChemicalsGCFunctions"
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
		
		http.Handle("/send-mobile-message", http.HandlerFunc(function.SendMobileMessage))
			
		log.Print("send-mobile-message started at: 3006")
		err = http.ListenAndServe(":3006", nil)
		if err != nil{
			log.Printf("Error occurred when starting the server: %v", err)
		} 
	}
}