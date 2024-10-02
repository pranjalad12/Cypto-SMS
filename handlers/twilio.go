package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"crypto-sms/services"
	"crypto-sms/utils"
)

type ParsedSMS struct {
	RecipientAddress string  `json:"recipient_address"`
	RecipientCrypto  string  `json:"recipient_crypto"`
	AmountUSD        float64 `json:"amount_usd"`
	Crypto           string  `json:"crypto"`
	Passkey          string  `json:"passkey"`
	Checksum         string  `json:"checksum"`
}

// HandleTwilioWebhook handles incoming SMS messages from Twilio
func HandleTwilioWebhook(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form data", http.StatusInternalServerError)
		return
	}

	from := r.FormValue("From")
	body := r.FormValue("Body")

	if from == "" || body == "" {
		http.Error(w, "Missing 'From' or 'Body' in form data", http.StatusBadRequest)
		return
	}

	// Ensure phone number has a plus sign
	if !strings.HasPrefix(from, "+") {
		from = "+" + from
	}

	// Parse the SMS content
	parsedSMS, err := ParseSMSContent(body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse SMS content: %v", err), http.StatusBadRequest)
		utils.SendSMS(from, os.Getenv("TWILIO_PHONE_NUMBER"), "Failed to parse SMS content")
		return
	}

	// Add phone number to parsed details
	transactionDetails := map[string]interface{}{
		"phone_number":       from,
		"recipient_address":  parsedSMS.RecipientAddress,
		"recipient_crypto":   parsedSMS.RecipientCrypto,
		"amount_usd":         parsedSMS.AmountUSD,
		"crypto":             parsedSMS.Crypto,
		"passkey":            parsedSMS.Passkey,
		"checksum":           parsedSMS.Checksum,
	}

	// Process the transaction
	err = services.ProcessTransaction(transactionDetails)
	if err != nil {
		http.Error(w, fmt.Sprintf("Transaction failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Send a response back to Twilio to acknowledge receipt of the message
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Message received"))
}

// ParseSMSContent parses the SMS content and extracts the transaction details
func ParseSMSContent(content string) (ParsedSMS, error) {
	sessionId, err := createSession()
	if err != nil {
		return ParsedSMS{}, fmt.Errorf("error creating session: %v", err)
	}

	query := fmt.Sprintf(`Parse the following SMS content and extract the following information:
    - Recipient Address
    - Recipient Crypto
    - Amount to send in USD
    - Cryptocurrency to use
    - Passkey
    - Checksum 

    SMS Content:
    %s

    Return the result as a JSON object with keys: recipient_address, recipient_crypto, amount_usd, crypto, passkey, checksum`, content)

	response, err := querySession(query, sessionId)
	if err != nil {
		return ParsedSMS{}, fmt.Errorf("error querying session: %v", err)
	}

	// Extract JSON from markdown code block
	jsonStart := strings.Index(response, "{")
	jsonEnd := strings.LastIndex(response, "}")
	if jsonStart == -1 || jsonEnd == -1 {
		return ParsedSMS{}, fmt.Errorf("could not find JSON in response")
	}
	jsonStr := response[jsonStart : jsonEnd+1]

	var parsedResult ParsedSMS
	err = json.Unmarshal([]byte(jsonStr), &parsedResult)
	if err != nil {
		return ParsedSMS{}, fmt.Errorf("error unmarshaling parsed result: %v", err)
	}

	return parsedResult, nil
}

// Dummy functions for session management
func createSession() (string, error) {
	return "dummySessionId", nil
}

func querySession(query, sessionId string) (string, error) {
	// Simulate a response for now
	simulatedResponse := `
	{
		"recipient_address": "0x930e4763495a0e962626Ae4Ca485Dd3FBef9Aa76",
		"recipient_crypto": "BTC",
		"amount_usd": 100,
		"crypto": "ETH",
		"passkey": "mySecurePasskey123",
		"checksum": "abc123def456"
	}`
	return simulatedResponse, nil
}
