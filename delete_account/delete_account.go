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
        shared.WriteJSONError(response, http.StatusMethodNotAllowed, "Wrong HTTP method")
        return
    }

    if err := shared.IsAuthorized(request); err != nil {
        shared.WriteJSONError(response, http.StatusUnauthorized, err.Error())
        return
    }

    defer request.Body.Close()

    uid := request.URL.Query().Get("uid")
    if uid == "" {
        shared.WriteJSONError(response, http.StatusBadRequest, "Missing uid parameter")
        return
    }

    if err := shared.AuthClient.DeleteUser(ctx, uid); err != nil {
        log.Printf("Auth delete error: %v", err)

        firebaseError := shared.ExtractFirebaseErrorFromResponse(err)
        if firebaseError.Error.Message != "" {
            shared.WriteJSONError(response, http.StatusInternalServerError, firebaseError.Error.Message)
        } else {
            shared.WriteJSONError(response, http.StatusInternalServerError, err.Error())
        }
        return
    }

    if _, err := shared.FirestoreClient.Collection("users").Doc(uid).Delete(ctx); err != nil {
        log.Printf("Firestore delete error: %v", err)
        shared.WriteJSONError(response, http.StatusInternalServerError, err.Error())
        return
    }

    shared.WriteJSONSuccess(response, http.StatusOK, "Account deleted successfully", nil)
}