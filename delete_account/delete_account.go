package function

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/HarshMohanSason/AHSChemicalsGCShared/shared"
)

type DeleteAccountRequest struct {
    UID string 		`json:"uid"`
}

func init(){
    if os.Getenv("ENV") != "DEBUG"{
       shared.InitFirebaseProd(nil)
       functions.HTTP("delete-account", DeleteAccount)
    }
}

func DeleteAccount(response http.ResponseWriter, request *http.Request) {
    ctx := request.Context()

    if shared.CorsEnabledFunction(response, request) {
        return
    }

    if request.Method != http.MethodDelete {
        http.Error(response, "Wrong HTTP method", http.StatusMethodNotAllowed)
        return
    }
    
    if err := shared.IsAuthorized(request); err != nil {
        http.Error(response, err.Error(), http.StatusUnauthorized)
        return
    }

    defer request.Body.Close()

    var req DeleteAccountRequest
    if err := json.NewDecoder(request.Body).Decode(&req); err != nil {
        log.Printf("JSON decode error: %v", err)
        http.Error(response, "Invalid request body", http.StatusBadRequest)
        return
    }

    if req.UID == "" {
        log.Print("UID is missing in the request")
        http.Error(response, "User ID not provided", http.StatusBadRequest)
        return
    }

    if err := shared.AuthClient.DeleteUser(ctx, req.UID); err != nil {
        log.Printf("Auth delete error: %v", err)
        http.Error(response, "Failed to delete user", http.StatusInternalServerError)
        return
    }

    if _, err := shared.FirestoreClient.Collection("users").Doc(req.UID).Delete(ctx); err != nil {
        log.Printf("Firestore delete error: %v", err)
        http.Error(response, "Failed to delete user", http.StatusInternalServerError)
        return
    }

    response.WriteHeader(http.StatusOK)
    if _, err := response.Write([]byte("Account deleted successfully")); err != nil {
        log.Printf("Response write error: %v", err)
    }
}