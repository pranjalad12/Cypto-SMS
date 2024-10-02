package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"crypto-sms/storage"
	"crypto-sms/utils"
)

func Generate2FACode(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PhoneNumber   string `json:"phone_number"`
		WalletAddress string `json:"wallet_address"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	code := utils.Generate2FACode()
	err := storage.Generate2FACodeAndStore(r.Context(), req.PhoneNumber, code)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = utils.SendSMS(req.PhoneNumber, os.Getenv("TWILIO_PHONE_NUMBER"), fmt.Sprintf("Your 2FA code is: %s", code))
	if err != nil {
		http.Error(w, "Failed to send 2FA code", http.StatusInternalServerError)
		return
	}

	response := struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}{
		Status:  "success",
		Message: "2FA code sent successfully",
	}

	json.NewEncoder(w).Encode(response)
}

func Verify2FACode(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PhoneNumber   string `json:"phone_number"`
		WalletAddress string `json:"wallet_address"`
		Code          string `json:"code"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	codeMatches, err := storage.Verify2FACode(r.Context(), req.PhoneNumber, req.Code)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if !codeMatches {
		http.Error(w, "Invalid 2FA code", http.StatusUnauthorized)
		return
	}

	_, phoneExists, err := storage.CheckPhoneNumberExistsInSmsService(r.Context(), req.PhoneNumber)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if phoneExists {
		http.Error(w, "Phone number already linked to another account", http.StatusConflict)
		return
	}

	err = storage.UpdatePhoneNumber(r.Context(), req.WalletAddress, req.PhoneNumber)
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

func Generate2FAHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PhoneNumber  string `json:"phone_number"`
		WalletAddress string `json:"wallet_address"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Generate a 2FA code
	code := utils.Generate2FACode()

	// Store the 2FA code in the database
	err = storage.Store2FACode(context.TODO(), req.PhoneNumber, code)
	if err != nil {
		http.Error(w, "Failed to store 2FA code", http.StatusInternalServerError)
		return
	}

	// Send the 2FA code via SMS
	err = utils.SendSMS(req.PhoneNumber, os.Getenv("TWILIO_PHONE_NUMBER"), "Your 2FA code is: " + code)
	if err != nil {
		http.Error(w, "Failed to send 2FA code", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "2FA code sent"})
}