package tests

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/HarshMohanSason/AHSChemicalsGCShared/shared"
	"github.com/joho/godotenv"
)

func TestMain(m *testing.M){

	err := godotenv.Load("../../keys/.env")
	if err != nil{
		log.Printf("Error occurred loading the env file: %v", err)
	}

	adminSDKFilePath := fmt.Sprintf("../%s", os.Getenv("FIREBASE_CREDENTIALS_DEBUG"))
	
	//Initialize the debug project sdk
	shared.InitFirebaseDebug(adminSDKFilePath)

	exitCode := m.Run()

	os.Exit(exitCode)
}