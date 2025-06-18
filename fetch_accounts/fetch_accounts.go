package function

import (
	"log"
	"net/http"
	"os"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/HarshMohanSason/AHSChemicalsGCShared/shared"
)

func init() {
	if os.Getenv("ENV") != "DEBUG" {
		shared.InitFirebaseProd(nil)
		functions.HTTP("fetch-accounts", FetchAccounts)
	}
}

// FetchAccounts fetches the first 1000 manager accounts. Will add pagination later if needed.
func FetchAccounts(response http.ResponseWriter, request *http.Request) {
	ctx := request.Context()

	if shared.CorsEnabledFunction(response, request) {
		return
	}

	if request.Method != http.MethodGet {
		shared.WriteJSONError(response, http.StatusMethodNotAllowed, "Wrong HTTP method")
		return
	}

	if err := shared.IsAuthorized(request); err != nil {
		shared.WriteJSONError(response, http.StatusUnauthorized, err.Error())
		return
	}

	iter := shared.AuthClient.Users(ctx, "")
	users := []map[string]any{}

	for {
		userRecord, err := iter.Next()
		if userRecord == nil {
			break
		}
		if err != nil {
			log.Printf("Auth iteration error: %v", err)

			firebaseError := shared.ExtractFirebaseErrorFromResponse(err)
			if firebaseError != nil {
				shared.WriteJSONError(response, http.StatusInternalServerError, firebaseError.Error.Message)
			} else {
				shared.WriteJSONError(response, http.StatusInternalServerError, err.Error())
			}
			return
		}

		// Only fetching customer accounts
		if userRecord.CustomClaims == nil {
			docSnapshot, err := shared.FirestoreClient.Collection("users").Doc(userRecord.UserRecord.UserInfo.UID).Get(ctx)
			if err != nil {
				log.Printf("Firestore read error: %v", err)
				shared.WriteJSONError(response, http.StatusInternalServerError, err.Error())
				return
			}

			data := docSnapshot.Data()
			properties := data["properties"]
			brands := data["brands"]

			userMap := map[string]any{
				"uid":         userRecord.UserRecord.UserInfo.UID,
				"displayName": userRecord.UserRecord.UserInfo.DisplayName,
				"email":       userRecord.UserRecord.UserInfo.Email,
				"properties":  properties,
				"brands":      brands,
			}
			users = append(users, userMap)
		}
	}

	shared.WriteJSONSuccess(response, http.StatusOK, "Fetched accounts successfully", users)
}
