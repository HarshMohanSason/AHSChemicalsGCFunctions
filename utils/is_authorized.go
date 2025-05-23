package utils

import (
	"errors"
	"net/http"
	"os"
)

//Only using this in production.
func IsAuthorized(request *http.Request) error {
		
		//True in local dev
		if os.Getenv("ENV") == "DEV"{
			return nil
		}

		ctx := request.Context()
			
		//Check if the cookie named session exists. 
		tokenCookie, err := request.Cookie("session")
		if err != nil{
			return err
		}

		idToken := tokenCookie.Value 
		token, err := AuthClient.VerifyIDToken(ctx, idToken)
		if err != nil {
			return err
		}
		if !token.Claims["isAdmin"].(bool){
			return errors.New("Unauthroized. Only admins are allowed to perform this operation")
		}

		return nil
}