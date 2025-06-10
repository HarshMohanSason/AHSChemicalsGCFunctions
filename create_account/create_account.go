package function

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"firebase.google.com/go/v4/auth"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/HarshMohanSason/AHSChemicalsGCShared/shared"
)

func init(){
    if os.Getenv("ENV") != "DEBUG"{
	   shared.InitFirebaseProd(nil)
	   functions.HTTP("create-account", CreateAccount)
    }
}

type CreateAccountRequest struct {
    Name        string      `json:"name"`
    Email      	string  	`json:"email"`
    Password   	string  	`json:"password"`
    Properties 	[]string	`json:"properties"`
    Brands     	[]string	`json:"brands"`
}

func CreateAccount(response http.ResponseWriter, request *http.Request) {
    ctx := request.Context()

    if shared.CorsEnabledFunction(response, request) {
        return
    }

    if request.Method != http.MethodPost {
        http.Error(response, "Wrong http method", http.StatusMethodNotAllowed)
        return
    }

    if err := shared.IsAuthorized(request); err != nil {
        http.Error(response, err.Error(), http.StatusUnauthorized)
        return
    }

    defer request.Body.Close()

    var req CreateAccountRequest
    if err := json.NewDecoder(request.Body).Decode(&req); err != nil {
        log.Printf("JSON decode error: %v", err)
        http.Error(response, "Invalid request body", http.StatusBadRequest)
        return
    }

    if len(req.Properties) == 0 {
        http.Error(response, "At least one property is required", http.StatusBadRequest)
        return
    }
    if len(req.Brands) == 0 {
        http.Error(response, "At least one brand is required", http.StatusBadRequest)
        return
    }

    userToCreate := (&auth.UserToCreate{}).
        Email(req.Email).
        Password(req.Password).
        DisplayName(req.Name)

    createdUser, err := shared.AuthClient.CreateUser(ctx, userToCreate)
    if err != nil {
        log.Printf("Auth create error: %v", err)
        http.Error(response, err.Error(), http.StatusInternalServerError)
        return
    }

    data := map[string]any{
        "properties": req.Properties,
        "brands":     req.Brands,
    }

    if _, err := shared.FirestoreClient.Collection("users").Doc(createdUser.UID).Set(ctx, data); err != nil {
        log.Printf("Firestore write error: %v", err)
        http.Error(response, err.Error(), http.StatusInternalServerError)
        return
    }

    response.WriteHeader(http.StatusOK)
    if _, err := response.Write([]byte(`Account Created Successfully!`)); err != nil {
        log.Printf("Response write error: %v", err)
    }
}
