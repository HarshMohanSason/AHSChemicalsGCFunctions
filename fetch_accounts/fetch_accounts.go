package function

import (
	"encoding/json"
	"net/http"

	"firebase.google.com/go/v4/auth"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/HarshMohanSason/AHSChemicalsGCShared/shared"
)

func init(){
	shared.InitFirebaseProd(nil)
	functions.HTTP("fetch-accounts", FetchAccounts)
}

//FetchAccounts only fetches a max of first 1000 users. Will update if users increase.  
func FetchAccounts(response http.ResponseWriter, request *http.Request){

	ctx := request.Context()
	
	if request.Method != http.MethodGet{
		http.Error(response, "Wrong http method", http.StatusMethodNotAllowed)
		return
	}

	if shared.CorsEnabledFunction(response, request){
		return
	}

	err := shared.IsAuthorized(request)
	if err != nil {
		http.Error(response, err.Error(), http.StatusUnauthorized)
		return
	}
 
	iter := shared.AuthClient.Users(ctx, "")
	users := []auth.UserRecord{}
	for {
		userRecord, err := iter.Next()
		if userRecord == nil{
			break
		}
		if err != nil{
			http.Error(response, err.Error(), http.StatusInternalServerError)
			break
		}
		users = append(users, *userRecord.UserRecord)
	}

	response.WriteHeader(http.StatusOK)
	err = json.NewEncoder(response).Encode(users)
	if err != nil {
    	http.Error(response, err.Error(), http.StatusInternalServerError)
    }
}