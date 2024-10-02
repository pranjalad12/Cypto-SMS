package handlers

import (
	"encoding/json"
	"net/http"
    "context"
	"os"
	"crypto-sms/storage"
	"crypto-sms/utils"
	
)

func CheckSMSServiceExists(w http.ResponseWriter, r *http.Request) {
	var req struct {
		WalletAddress string `json:"wallet_address"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	service, exists, err := storage.CheckWalletExistsInSmsService(r.Context(), req.WalletAddress)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := struct {
		DoesExist bool    `json:"does_exist"`
		IsPrimary bool    `json:"is_primary,omitempty"`
		Passkey   string  `json:"passkey,omitempty"`
		Limit     float64 `json:"limit,omitempty"`
	}{
		DoesExist: exists,
	}

	if exists {
		if service.PhoneNumber == "" {
			response.IsPrimary = false
		} else {
			response.IsPrimary = true
			response.Passkey = service.Passkey
			response.Limit = service.Limit
		}
	}

	json.NewEncoder(w).Encode(response)
}

func CreateSMSService(w http.ResponseWriter, r *http.Request) {
	var req struct {
		WalletAddress string `json:"wallet_address"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	privateKey, publicKey, err := utils.GenerateKeyPair()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	service := storage.SmsService{
		WalletAddress: req.WalletAddress,
		PublicKey:     publicKey,
		Limit:         1000,
	}
	err = storage.CreateSmsService(r.Context(), service)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := struct {
		Status     string `json:"status"`
		PrivateKey string `json:"private_key"`
	}{
		Status:     "created",
		PrivateKey: privateKey,
	}

	json.NewEncoder(w).Encode(response)
}

func UpdateSmsService(w http.ResponseWriter, r *http.Request) {
	var req struct {
		WalletAddress string  `json:"wallet_address"`
		Passkey       string  `json:"passkey"`
		Limit         float64 `json:"limit"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	err := storage.UpdateSmsService(r.Context(), req.WalletAddress, req.Passkey, req.Limit)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}{
		Status:  "success",
		Message: "SMS service updated successfully",
	}

	json.NewEncoder(w).Encode(response)
}

func UpdatePhoneNumber(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PhoneNumber   string `json:"phone_number"`
		WalletAddress string `json:"wallet_address"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	err := storage.UpdatePhoneNumber(r.Context(), req.WalletAddress, req.PhoneNumber)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}{
		Status:  "success",
		Message: "Phone number updated successfully",
	}

	json.NewEncoder(w).Encode(response)
}

func SendDummySMS(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PhoneNumber string `json:"phone_number"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	err := utils.SendSMS(req.PhoneNumber, os.Getenv("TWILIO_PHONE_NUMBER"), "This is a test message from CryptoSMS.")
	if err != nil {
		http.Error(w, "Failed to send SMS", http.StatusInternalServerError)
		return
	}

	response := struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}{
		Status:  "success",
		Message: "Dummy SMS sent successfully",
	}

	json.NewEncoder(w).Encode(response)
}
func ListSmsServicesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.TODO()
	services, err := storage.ListSmsServices(ctx)
	if err != nil {
		http.Error(w, "Failed to list SMS services", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(services)
}

func UpdatePhoneNumberHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PhoneNumber  string `json:"phone_number"`
		WalletAddress string `json:"wallet_address"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Update the phone number in the database
	err = storage.UpdatePhoneNumber(context.TODO(), req.WalletAddress, req.PhoneNumber)
	if err != nil {
		http.Error(w, "Failed to update phone number", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "Phone number updated"})
}
