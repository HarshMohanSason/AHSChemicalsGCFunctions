package function

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/HarshMohanSason/AHSChemicalsGCShared/shared"
	"github.com/HarshMohanSason/AHSChemicalsGCShared/shared/cors"
	firebase_shared "github.com/HarshMohanSason/AHSChemicalsGCShared/shared/firebase"
	"github.com/HarshMohanSason/AHSChemicalsGCShared/shared/quickbooks"
	"github.com/HarshMohanSason/AHSChemicalsGCShared/shared/utils"
)

func init() {
	if os.Getenv("ENV") != "DEBUG" {
		ctx := context.Background()
		quickbooks.InitQuickBooksProd(ctx)
		firebase_shared.InitFirebaseProd(nil)

		functions.HTTP("quickbooks-create-invoice", CreateInvoice)
	}
}

func CreateInvoice(response http.ResponseWriter, request *http.Request) {
	ctx := request.Context()

	// Handle CORS preflight requests
	if cors.CorsEnabledFunction(response, request) {
		return
	}

	// Ensure method is POST
	if request.Method != http.MethodPost {
		firebase_shared.WriteJSONError(response, http.StatusMethodNotAllowed, "Expected POST request")
		return
	}

	// Authenticate Firebase admin user
	uid, err := firebase_shared.GetUIDIfAdmin(request)
	if err != nil {
		firebase_shared.WriteJSONError(response, http.StatusUnauthorized, err.Error())
		return
	}

	// Get valid QuickBooks access token
	tokenData, err := quickbooks.EnsureValidAccessToken(ctx, uid)
	if err != nil {
		firebase_shared.WriteJSONError(response, http.StatusUnauthorized, "QuickBooks authorization invalid or expired. Please authenticate again.")
		return
	}

	// Read and parse the request body
	defer request.Body.Close()
	body, err := io.ReadAll(request.Body)
	if err != nil {
		firebase_shared..WriteJSONError(response, http.StatusBadRequest, "Error reading request body: "+err.Error())
		return
	}

	var invoicePayload map[string]any
	if err := json.Unmarshal(body, &invoicePayload); err != nil {
		firebase_shared.WriteJSONError(response, http.StatusBadRequest, err.Error())
		return
	}

	// Build the QuickBooks API request
	apiURL := fmt.Sprintf("%s/v3/company/%s/invoice", quickbooks.QUICKBOOKS_API_URL, realmID)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewBuffer(body))
	if err != nil {
		shared.WriteJSONError(response, http.StatusInternalServerError, "Error creating QuickBooks request: "+err.Error())
		return
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokenData["access_token"].(string)))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	// Execute the API request
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		shared.WriteJSONError(response, http.StatusInternalServerError, "Error sending request to QuickBooks API: "+err.Error())
		return
	}
	defer resp.Body.Close()

	// Read and parse QuickBooks response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		shared.WriteJSONError(response, http.StatusInternalServerError, "Error reading QuickBooks response: "+err.Error())
		return
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		shared.WriteJSONError(response, resp.StatusCode, fmt.Sprintf("QuickBooks API Error: %s", string(respBody)))
		return
	}

	var invoiceResponse map[string]any
	if err := json.Unmarshal(respBody, &invoiceResponse); err != nil {
		shared.WriteJSONError(response, http.StatusInternalServerError, "Error parsing QuickBooks response JSON: "+err.Error())
		return
	}

	shared.WriteJSONSuccess(response, http.StatusOK, "Invoice created successfully", invoiceResponse)
}
