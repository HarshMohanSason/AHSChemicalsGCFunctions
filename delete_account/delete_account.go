package function

import (
    "log"
    "net/http"
    "os"

    "github.com/GoogleCloudPlatform/functions-framework-go/functions"
    "github.com/HarshMohanSason/AHSChemicalsGCShared/shared"
)

func init() {
    // Register the delete-account function for production environment only
    if os.Getenv("ENV") != "DEBUG" {
        shared.InitFirebaseProd(nil)
        functions.HTTP("delete-account", DeleteAccount)
    }
}

// DeleteAccount deletes a user from Firebase Authentication and removes their Firestore record.
//
// Authorization: Requires a valid Bearer token with 'admin' custom claim set to true.
// Method: DELETE
// URL Parameters:
//   - uid: The UID of the user to delete (required)
//
// Success Response: 200 OK with success message
// Error Response: Appropriate HTTP status codes with descriptive error messages
func DeleteAccount(response http.ResponseWriter, request *http.Request) {
    ctx := request.Context()

    // Handle CORS (OPTIONS) requests and setup CORS headers
    if shared.CorsEnabledFunction(response, request) {
        return
    }

    // Only allow DELETE method
    if request.Method != http.MethodDelete {
        shared.WriteJSONError(response, http.StatusMethodNotAllowed, "Wrong HTTP method")
        return
    }

    // Ensure that the user is authorized (admin privileges)
    if err := shared.IsAuthorizedAndAdmin(request); err != nil {
        shared.WriteJSONError(response, http.StatusUnauthorized, err.Error())
        return
    }

    defer request.Body.Close()

    // Get UID from query parameter
    uid := request.URL.Query().Get("uid")
    if uid == "" {
        shared.WriteJSONError(response, http.StatusBadRequest, "Missing uid parameter")
        return
    }

    // Delete user from Firebase Authentication
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

    // Delete associated Firestore document
    if _, err := shared.FirestoreClient.Collection("users").Doc(uid).Delete(ctx); err != nil {
        log.Printf("Firestore delete error: %v", err)
        shared.WriteJSONError(response, http.StatusInternalServerError, err.Error())
        return
    }

    // Respond with success
    shared.WriteJSONSuccess(response, http.StatusOK, "Account deleted successfully", nil)
}