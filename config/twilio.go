package config

import (
	"os"

	"github.com/twilio/twilio-go"
)

func TwilioClient() (*twilio.RestClient, string, error) {

	serviceSID := os.Getenv("TWILIO_SERVICE_SID")
	authToken := os.Getenv("TWILIO_AUTH_TOKEN")
	accountSID := os.Getenv("TWILIO_ACCOUNT_SID")

	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: accountSID,
		Password: authToken,
	})
	return client, serviceSID, nil
}
