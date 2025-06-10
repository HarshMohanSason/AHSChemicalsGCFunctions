package function

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/HarshMohanSason/AHSChemicalsGCShared/shared"
	sendgrid "github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type EmailMetaData struct{
	Recipients    map[string]string   		`json:"recipients"`
	Data 		  map[string]any			`json:"data"`
	TemplateID    string                    `json:"template_id"`
}

func Init(){
	if os.Getenv("ENV") != "DEBUG"{
		functions.HTTP("send-mail", SendMail)
	}
}

func SendMail(response http.ResponseWriter, request *http.Request){

	if shared.CorsEnabledFunction(response, request){
		return
	}

	if request.Method != http.MethodPost{
		http.Error(response, "Wrong http method", http.StatusMethodNotAllowed)
		return
	}

	var emailMetaData EmailMetaData
	if err:= json.NewDecoder(request.Body).Decode(&emailMetaData); err != nil{
		log.Printf("Error occured decoding the email: %v", err)
		http.Error(response, "Error in retreving the sent data", http.StatusBadRequest)
		return
	}	

	from := mail.NewEmail("AHSChemicals", "inbox@azurehospitalitysupply.com")

 	var recipents []*mail.Email
 	for email, name := range emailMetaData.Recipients {
 	 	recipents = append(recipents, mail.NewEmail(name, email))
 	}	

	// Personalization for the mail
	p := mail.NewPersonalization()
	
	p.AddTos(recipents...)
	for key, value := range emailMetaData.Data{
		p.SetDynamicTemplateData(key, value)
	}
	
	//Create a new mail instance
	message := mail.NewV3Mail()
	message.SetFrom(from)
	message.AddPersonalizations(p)

	//Template ID
	message.SetTemplateID(emailMetaData.TemplateID)

	//Create a new client
	client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API_KEY"))
	_, err := client.Send(message)

	if err != nil{
		log.Printf("Error sending the mail to the recipents %v", err)
	}

	//Not sending any status, mail is sent silently
}