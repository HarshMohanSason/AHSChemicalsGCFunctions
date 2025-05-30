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
	UserEmail	  string     `json:"userEmail"` //User who placed the order
	UserUID		  string     `json:"userUID"`
	OrderID 	  string  	 `json:"orderID"`
	Subject   	  string     `json:"emailSubject"`
	Content 	  string 	 `json:"emailBody"`
	Attachments	  []string   `json:"attachments"`
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

	var emailMetaData EmailMetaData
	if err:= json.NewDecoder(request.Body).Decode(&emailMetaData); err != nil{
		log.Printf("Error occured decoding the email: %v", err)
		http.Error(response, "Error in retreving the sent data", http.StatusBadRequest)
		return
	}	

	//TODO: Need to set the proper recipents email, set the html content and the attachments
	from := mail.NewEmail("AHSChemicals", "azurehospitalitysupply.com.")
	to1 := mail.NewEmail("Recipient One", "recipient1@example.com")
	to2 := mail.NewEmail("Recipient Two", "recipient2@example.com")
	to3 := mail.NewEmail("Recipient Two", "recipient2@example.com")

	// Personalization for the mail
	p := mail.NewPersonalization()
	p.AddTos(to1, to2, to3)
	p.Subject = emailMetaData.Subject

	//Create a new mail instance
	message := mail.NewV3Mail()
	message.SetFrom(from)
	message.AddPersonlizations(p)

	//Add the body content 
	content := mail.NewContent("text/html", emailMetaData.Content)
	message.AddContent(content) 

	//Create a new client
	client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API_KEY"))
	_, err := client.Send(message)


	if err != nil{
		log.Printf("Error sending the mail to the recipents %v", err)
	}
}