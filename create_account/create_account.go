package function

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"firebase.google.com/go/v4/auth"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/HarshMohanSason/AHSChemicalsGCShared/shared"
)

func init(){
	shared.InitFirebaseProd(nil)
	functions.HTTP("create-account", CreateAccount)
}

func CreateAccount(response http.ResponseWriter, request *http.Request) {

	ctx := request.Context()

	if request.Method != http.MethodPost{
		http.Error(response, "Wrong http method", http.StatusMethodNotAllowed)
		return
	}	
	//CORS 
	if shared.CorsEnabledFunction(response, request){
		return
	}

	//Check if the current user is an admin or not
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

	//Filter the data
	data, err := extractUserData(user)
	if err != nil{
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

	//Add properties to firestore
	_, err = shared.FirestoreClient.Collection("users").Doc(createdUser.UID).Set(ctx, data)

	if err != nil {
		log.Print(err)
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}

	response.WriteHeader(http.StatusOK)
	response.Write([]byte(`{"message":"Account created successfully"}`))
}

func extractUserData(user map[string]interface{}) (map[string]interface{}, error){
		
	if user == nil{
		return nil, errors.New("User cannot be nil")
	}

	properties, ok := (user)["properties"].([]interface{})
	if !ok || len(properties) == 0{
		return nil, errors.New("Need to have at least one property for the user")
	}

	brands, ok := (user)["brands"].([]interface{})
	if !ok || len(brands) == 0{
		return nil, errors.New("Need to select at least one brand ")
	}

	data := map[string]interface{}{
		"properties": properties,
		"brands": brands,
	}

	return data, nil
}
