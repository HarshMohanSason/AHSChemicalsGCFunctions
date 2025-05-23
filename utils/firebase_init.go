package utils

import (
	"context"
	"log"
	"os"
	"sync"
	"github.com/joho/godotenv"
	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

var (
	App              *firebase.App
	AuthClient       *auth.Client
	FirestoreClient  *firestore.Client
	initOnce         sync.Once
)

func InitFirebase(){
	initOnce.Do(func(){
		ctx := context.Background()
		var err error
		
		//Load the env file
		err = godotenv.Load("keys/.env") 
		if err != nil {
			log.Fatalf("Unable to load the env file %v", err)
		} 
		
		//Get the firebase credentials
		cred := os.Getenv("FIREBASE_CREDENTIALS") 
		
		//Make sure credentials fetched are valid
		if cred != "" {
			opt := option.WithCredentialsFile(cred)
			
			//Initialize the firebase app
			App, err = firebase.NewApp(ctx, nil, opt)
			if err != nil{
				log.Fatalf("Unable to register the app: %v", err)
			}
		} else {
			log.Fatal("Credentials path is empty.")
		}

		//Once thbe app is initialized, initialize the AppClient
		AuthClient, err = App.Auth(ctx)
		if err != nil {
			log.Fatalf("Failed to initialize Auth client: %v", err)
		}

		//Initialize the firestore
		FirestoreClient, err = App.Firestore(ctx)
		if err != nil{
			log.Fatalf("Failed to initialize Firestore client: %v", err)
		}
	})
}