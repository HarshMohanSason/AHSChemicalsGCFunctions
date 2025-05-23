package createaccount

import (
	"encoding/json"
	"firebase.google.com/go/v4/auth"
	"fmt"
	"io"
	"log"
	"net/http"
	"shared"
)

func CreateAccount(response http.ResponseWriter, request *http.Request) {
	ctx := request.Context()

	if shared.HandleCors(response, request) {
		return
	}

	//Handle the Auth for the request now
	err := shared.IsAuthorized(request)
	if err != nil {
		http.Error(response, err.Error(), http.StatusUnauthorized)
		return
	}

	bodyBytes, err := io.ReadAll(request.Body)
	if err != nil {
		log.Printf("Error reading the body: %v", err)
		http.Error(response, err.Error(), http.StatusBadRequest)
		return
	}

	var user map[string]interface{}

	err = json.Unmarshal(bodyBytes, &user)
	if err != nil {
		log.Print(err)
		http.Error(response, err.Error(), http.StatusBadRequest)
		return
	}

	displayName := fmt.Sprintf("%s %s", user["firstName"], user["lastName"])

	//Create the user object
	userToCreate := (&auth.UserToCreate{}).
		Email(user["email"].(string)).
		Password(user["password"].(string)).
		DisplayName(displayName)

	//Create the user
	createdUser, err := shared.AuthClient.CreateUser(
		ctx, userToCreate)

	if err != nil {
		log.Print(err)
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"properties": user["properties"],
	}
	if data["properties"] == nil {
		http.Error(response, "Properties cannot be empty", http.StatusInternalServerError)
		return
	}

	//Add the data to firestore
	_, err = shared.FirestoreClient.Collection("users").Doc(createdUser.UID).Set(ctx, data)

	if err != nil {
		log.Print(err)
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}

	response.WriteHeader(http.StatusOK)
	response.Write([]byte(`{"message":"Account created successfully"}`))
}
