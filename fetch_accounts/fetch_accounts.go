package function

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/HarshMohanSason/AHSChemicalsGCShared/shared"
)

func init(){
    if os.Getenv("ENV") != "DEBUG"{
	   shared.InitFirebaseProd(nil)
	   functions.HTTP("fetch-accounts", FetchAccounts)
    }
}

//FetchAccounts only fetches the first 1000 managers accounts. Will update if users increase
func FetchAccounts(response http.ResponseWriter, request *http.Request){

	ctx := request.Context()

	if shared.CorsEnabledFunction(response, request){
		return
	}

	if request.Method != http.MethodGet{
		log.Print("Wrong http method")
		http.Error(response, "Wrong http method", http.StatusMethodNotAllowed)
		return
	}

	err := shared.IsAuthorized(request)
	if err != nil {
		log.Print(err.Error())
		http.Error(response, err.Error(), http.StatusUnauthorized)
		return
	}
 
	iter := shared.AuthClient.Users(ctx, "")
	users := []map[string]interface{}{}
	for {
		userRecord, err := iter.Next()
		if userRecord == nil{
			break
		}
		if err != nil{
			log.Print(err.Error())
			http.Error(response, err.Error(), http.StatusInternalServerError)
			return
		}
		//Only fetching the managers not the admins since managers do not have any custom claims set
		if userRecord.CustomClaims == nil{
			docSnapshot, err := shared.FirestoreClient.Collection("users").Doc(userRecord.UserRecord.UserInfo.UID).Get(ctx)
			
			if err != nil{
				log.Print(err.Error())
				http.Error(response, err.Error(),  http.StatusInternalServerError)
				return
			}
	
			data := docSnapshot.Data()
			properties:= data["properties"]	
        	brands := data["brands"]
	
			userMap := map[string]interface{}{
				"uid": userRecord.UserRecord.UserInfo.UID,
				"displayName": userRecord.UserRecord.UserInfo.DisplayName,
				"email":  userRecord.UserRecord.UserInfo.Email,
				"properties": properties, 
				"brands": brands,
			}
			users = append(users, userMap)
		}
	}	
	
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusOK)
	err = json.NewEncoder(response).Encode(users)
	if err != nil {
		log.Print(err.Error())
    	http.Error(response, err.Error(), http.StatusInternalServerError)
    }
}
