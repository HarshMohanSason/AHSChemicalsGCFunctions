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

	//For some reason go tests run one dir down, so moving up one dir
	dirPath := "../" 
	envPath := "../keys/.env"
	pathToLoad := fmt.Sprintf("%s%s", dirPath, envPath)
	
	err := godotenv.Load(pathToLoad)
	if err != nil{
		log.Printf("Error occurred loading the env file: %v", err)
	}

	adminSDKFilePath := fmt.Sprintf("%s%s", dirPath, os.Getenv("FIREBASE_CREDENTIALS_DEBUG"))

	//Initialize the debug project sdk
	shared.InitFirebaseDebug(adminSDKFilePath)

	exitCode := m.Run()

	os.Exit(exitCode)
}