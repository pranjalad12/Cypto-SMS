package main

import (
	"log"
	"net/http"
	"os"

	"crypto-sms/handlers"
	"crypto-sms/storage"
)

func main() {
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		log.Fatal("MONGODB_URI environment variable is required")
	}
	storage.InitMongoDB(mongoURI)

	http.HandleFunc("/twilio-webhook", handlers.HandleTwilioWebhook)
	http.HandleFunc("/check-sms-service", handlers.CheckSMSServiceExists)
	http.HandleFunc("/create-sms-service", handlers.CreateSMSService)
	http.HandleFunc("/generate-2fa-code", handlers.Generate2FACode)
	http.HandleFunc("/verify-2fa-code", handlers.Verify2FACode)
	http.HandleFunc("/update-phone-number", handlers.UpdatePhoneNumber)
	http.HandleFunc("/update-sms-service", handlers.UpdateSmsService)
	http.HandleFunc("/send-dummy-sms", handlers.SendDummySMS)

	log.Println("HTTP server listening on port 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
