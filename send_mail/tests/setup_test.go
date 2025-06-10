package tests

import (
	"log"
	"os"
	"testing"

	"github.com/joho/godotenv"
)

func TestMain(m *testing.M){

	err := godotenv.Load("../../keys/.env")
	if err != nil{
		log.Printf("Error occurred loading the env file: %v", err)
	}

	exitCode := m.Run()

	os.Exit(exitCode)
}