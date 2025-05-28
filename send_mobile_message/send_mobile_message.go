package function

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/HarshMohanSason/AHSChemicalsGCShared/shared"
	twilio "github.com/twilio/twilio-go"
	twilioApi "github.com/twilio/twilio-go/rest/api/v2010"
)

type ReceivedData struct{
	UID			  string 		`json:"uid"` //This UID is of the user who placed the order
	Message   	  string 		`json:"message"`
}

func init(){
	functions.HTTP("send-mobile-message", SendMobileMessage)
}

func SendMobileMessage(response http.ResponseWriter, request *http.Request){

	if shared.CorsEnabledFunction(response, request){
		return
	}

	if request.Method != http.MethodPost{
		log.Print("Wrong http method")
		http.Error(response, "Wrong http method", http.StatusMethodNotAllowed)
		return
	}

	var data ReceivedData
	if err := json.NewDecoder(request.Body).Decode(&data); err != nil{
		log.Print(err)
		http.Error(response, "Error decoding the message:", http.StatusBadRequest)
		return
	}	

	//To phones are a string of multiple numbers separated by a comma
	recipents := os.Getenv("TWILIO_RECIPENTS_PHONE")	
	toPhones := strings.Split(recipents, ",")
	
	accountSid := os.Getenv("TWILIO_ACCOUNT_SID") 
	authToken := os.Getenv("TWILIO_AUTH_TOKEN")   
	fromPhone := os.Getenv("TWILIO_FROM_PHONE")   

	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: accountSid,
		Password: authToken,
	})	

	for _, receiverPhone := range toPhones{

		params := &twilioApi.CreateMessageParams{}
		params.SetTo(receiverPhone)
		params.SetFrom(fromPhone)
		params.SetBody(data.Message)
			
		//Message is sent silently to the admin
		_, err := client.Api.CreateMessage(params)
		if err != nil{
			log.Print("Error sending the message %v", err)
			continue
		}
	}

}