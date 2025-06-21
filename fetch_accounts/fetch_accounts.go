package function

import (
	"log"
	"net/http"
	"os"
	"sync"

	"firebase.google.com/go/v4/auth"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/HarshMohanSason/AHSChemicalsGCShared/shared"
)

func init() {
	// Initialize Firebase and register the Cloud Function only in production environment.
	if os.Getenv("ENV") != "DEBUG" {
		shared.InitFirebaseProd(nil)
		functions.HTTP("fetch-accounts", FetchAccounts)
	}
}

// FetchAccounts is a Google Cloud Function HTTP handler that retrieves all Firebase Auth users
// who do not have custom claims set, along with their associated Firestore document data (properties & brands).
//
// Authentication: Requires a valid Firebase ID token with 'admin' custom claim set to true in the Authorization header.
// Method: GET
// Response: JSON with a list of user data or error message
func FetchAccounts(response http.ResponseWriter, request *http.Request) {
	ctx := request.Context()

	// Handle CORS and preflight requests
	if shared.CorsEnabledFunction(response, request) {
		return
	}

	// Validate HTTP method
	if request.Method != http.MethodGet {
		shared.WriteJSONError(response, http.StatusMethodNotAllowed, "Wrong HTTP method")
		return
	}

	// Ensure that the user is authorized (admin privileges)
	if err := shared.IsAuthorizedAndAdmin(request); err != nil {
		shared.WriteJSONError(response, http.StatusUnauthorized, err.Error())
		return
	}

	iter := shared.AuthClient.Users(ctx, "") // Iterator to fetch all Firebase users
	users := []map[string]any{}
	var mu sync.Mutex // Protect concurrent writes to `users` slice

	const maxConcurrentLimit = 50
	var wg sync.WaitGroup
	sem := make(chan struct{}, maxConcurrentLimit)

	var errs []error
	var errsMu sync.Mutex // Protect concurrent writes to `errs` slice

	// Iterate through Firebase users
	for {
		userRecord, err := iter.Next()
		if userRecord == nil {
			break
		}
		if err != nil {
			log.Printf("Error fetching user records: %v", err)
			shared.WriteJSONError(response, http.StatusInternalServerError, err.Error())
			return
		}

		record := userRecord // Capture for use in the goroutine

		wg.Add(1)
		sem <- struct{}{} // Acquire a slot for concurrency limiting

		go func(record *auth.ExportedUserRecord) {
			defer wg.Done()
			defer func() { <-sem }() // Release slot when done

			// Process users who do not have custom claims set
			if record.CustomClaims == nil {
				docSnapshot, err := shared.FirestoreClient.Collection("users").Doc(record.UserRecord.UserInfo.UID).Get(ctx)
				if err != nil {
					errsMu.Lock()
					errs = append(errs, err)
					errsMu.Unlock()
					return
				}

				data := docSnapshot.Data()
				properties := data["properties"]
				brands := data["brands"]

				userMap := map[string]any{
					"uid":         record.UserRecord.UserInfo.UID,
					"displayName": record.UserRecord.UserInfo.DisplayName,
					"email":       record.UserRecord.UserInfo.Email,
					"properties":  properties,
					"brands":      brands,
				}

				mu.Lock()
				users = append(users, userMap)
				mu.Unlock()
			}
		}(record)
	}

	wg.Wait()

	// If any errors occurred while fetching Firestore data, respond with an error
	if len(errs) > 0 {
		log.Print("Error occurred while fetching some user accounts")
		shared.WriteJSONError(response, http.StatusInternalServerError, "Some user data could not be fetched")
		return
	}

	shared.WriteJSONSuccess(response, http.StatusOK, "Fetched accounts successfully", users)
}