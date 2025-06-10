package function

import (
	"log"
	"net/http"
	"os"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/HarshMohanSason/AHSChemicalsGCShared/shared"
)
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

    uid := request.URL.Query().Get("uid")
    if uid == "" {
        http.Error(response, "Missing uid parameter", http.StatusBadRequest)
        return
    }

    if err := shared.AuthClient.DeleteUser(ctx, uid); err != nil {
        log.Printf("Auth delete error: %v", err)
        http.Error(response, "Failed to delete user", http.StatusInternalServerError)
        return
    }

    if _, err := shared.FirestoreClient.Collection("users").Doc(uid).Delete(ctx); err != nil {
        log.Printf("Firestore delete error: %v", err)
        http.Error(response, "Failed to delete user", http.StatusInternalServerError)
        return
    }

    response.WriteHeader(http.StatusOK)
    if _, err := response.Write([]byte("Account deleted successfully")); err != nil {
        log.Printf("Response write error: %v", err)
    }
}