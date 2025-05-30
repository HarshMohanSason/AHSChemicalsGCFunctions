package function

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"reflect"

	"cloud.google.com/go/firestore"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/HarshMohanSason/AHSChemicalsGCShared/shared"
)

type UpdateUserRequest struct {
	UID        string   `json:"uid"`
	Brands     []string `json:"brands"`
	Properties []string `json:"properties"`
}

func init() {
	if os.Getenv("ENV") != "DEBUG"{
		shared.InitFirebaseProd(nil)
		functions.HTTP("update-account", UpdateAccount)
	}
}

func UpdateAccount(response http.ResponseWriter, request *http.Request) {
	ctx := request.Context()

	if shared.CorsEnabledFunction(response, request) {
		return
	}

	if err := shared.IsAuthorized(request); err != nil{
		http.Error(response, "Not authorized to do this action", http.StatusUnauthorized)
		return
	}

	if request.Method != http.MethodPut {
		log.Print("Wrong http method, expected method PUT")
		http.Error(response, "Error occurred, wrong http method", http.StatusMethodNotAllowed)
		return
	}

	defer request.Body.Close()

	var user UpdateUserRequest
	if err := json.NewDecoder(request.Body).Decode(&user); err != nil {
		log.Printf("Error occurred while decoding the body: %v", err)
		http.Error(response, "Error occurred while fetching the body", http.StatusBadRequest)
		return
	}

	if user.UID == "" {
		http.Error(response, "UID is required", http.StatusBadRequest)
		return
	}
	if len(user.Properties) == 0 {
		http.Error(response, "User Properties cannot be empty", http.StatusBadRequest)
		return
	}
	if len(user.Brands) == 0 {
		http.Error(response, "User Brands cannot be empty", http.StatusBadRequest)
		return
	}
	
	docSnapshot, err := shared.FirestoreClient.Collection("users").Doc(user.UID).Get(ctx)
	if err != nil {
		log.Printf("Error fetching documents for uid %v: %v", user.UID, err)
		http.Error(response, "No User found", http.StatusBadRequest)
		return
	}

	currentData := docSnapshot.Data()

	updates := []firestore.Update{}
	if !reflect.DeepEqual(currentData["brands"], user.Brands) {
		updates = append(updates, firestore.Update{FieldPath: []string{"brands"}, Value: user.Brands})
	}
	if !reflect.DeepEqual(currentData["properties"], user.Properties) {
		updates = append(updates, firestore.Update{FieldPath: []string{"properties"}, Value: user.Properties})
	}

	if len(updates) == 0 {
		response.WriteHeader(http.StatusOK)
		_, _ = response.Write([]byte("No changes detected"))
		return
	}

	_, err = shared.FirestoreClient.Collection("users").Doc(user.UID).Update(ctx, updates)
	if err != nil {
		log.Print(err)
		http.Error(response, "Error occurred while updating the user data", http.StatusInternalServerError)
		return
	}

	response.WriteHeader(http.StatusOK)
	_, err = response.Write([]byte("Information updated successfully"))
	if err != nil {
		log.Print(err)
		http.Error(response, "Error writing the response,", http.StatusInternalServerError)
	}
}