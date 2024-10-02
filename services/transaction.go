package services

import (
	"context"
	"fmt"
	"os"
	"strings"

	"crypto-sms/storage"
	"crypto-sms/utils"
)

// ProcessTransaction handles the transaction logic
func ProcessTransaction(details map[string]interface{}) error {
	ctx := context.TODO()
	phoneNumber := details["phone_number"].(string)
	passkey := details["passkey"].(string)
	amountUSD := details["amount_usd"].(float64)
	crypto := details["crypto"].(string)
	recipientAddress := details["recipient_address"].(string)
	recipientCrypto := details["recipient_crypto"].(string)

	// Ensure phone number has a plus sign
	if !strings.HasPrefix(phoneNumber, "+") {
		phoneNumber = "+" + phoneNumber
	}

	// Fetch sender's wallet address from sms_service
	senderService, exists, err := storage.CheckPhoneNumberExistsInSmsService(ctx, phoneNumber)
	if err != nil {
		utils.SendSMS(phoneNumber, os.Getenv("TWILIO_PHONE_NUMBER"), "Internal server error")
		return fmt.Errorf("error checking phone number: %w", err)
	}
	if !exists {
		utils.SendSMS(phoneNumber, os.Getenv("TWILIO_PHONE_NUMBER"), "Phone number not registered")
		return fmt.Errorf("phone number not registered")
	}
	if senderService.Passkey != passkey {
		utils.SendSMS(phoneNumber, os.Getenv("TWILIO_PHONE_NUMBER"), "Invalid passkey")
		return fmt.Errorf("invalid passkey")
	}
	if amountUSD > senderService.Limit {
		utils.SendSMS(phoneNumber, os.Getenv("TWILIO_PHONE_NUMBER"), "Transaction amount exceeds limit")
		return fmt.Errorf("transaction amount exceeds limit")
	}

	// Fetch sender's crypto balance from custodian
	senderCustodian, exists, err := storage.GetCustodianByWalletAddress(ctx, senderService.WalletAddress)
	if err != nil {
		utils.SendSMS(phoneNumber, os.Getenv("TWILIO_PHONE_NUMBER"), "Internal server error")
		return fmt.Errorf("error fetching custodian data: %w", err)
	}
	if !exists || senderCustodian.Cryptocurrencies[crypto] < amountUSD {
		utils.SendSMS(phoneNumber, os.Getenv("TWILIO_PHONE_NUMBER"), fmt.Sprintf("Insufficient %s balance", crypto))
		return fmt.Errorf("insufficient %s balance", crypto)
	}

	// Fetch recipient's custodian data
	recipientCustodian, exists, err := storage.GetCustodianByWalletAddress(ctx, recipientAddress)
	if err != nil {
		utils.SendSMS(phoneNumber, os.Getenv("TWILIO_PHONE_NUMBER"), "Internal server error")
		return fmt.Errorf("error fetching recipient custodian data: %w", err)
	}
	if !exists {
		recipientCustodian = &storage.Custodian{
			WalletAddress:    recipientAddress,
			Cryptocurrencies: make(map[string]float64),
		}
	}

	// Perform the transaction
	senderCustodian.Cryptocurrencies[crypto] -= amountUSD
	recipientCustodian.Cryptocurrencies[recipientCrypto] += amountUSD

	// Update custodian data in the database
	err = storage.UpdateCustodian(ctx, senderCustodian)
	if err != nil {
		utils.SendSMS(phoneNumber, os.Getenv("TWILIO_PHONE_NUMBER"), "Internal server error")
		return fmt.Errorf("error updating sender custodian data: %w", err)
	}
	err = storage.UpdateCustodian(ctx, recipientCustodian)
	if err != nil {
		utils.SendSMS(phoneNumber, os.Getenv("TWILIO_PHONE_NUMBER"), "Internal server error")
		return fmt.Errorf("error updating recipient custodian data: %w", err)
	}

	// Fetch recipient's phone number from sms_service
	recipientService, exists, err := storage.CheckWalletExistsInSmsService(ctx, recipientAddress)
	if err != nil {
		utils.SendSMS(phoneNumber, os.Getenv("TWILIO_PHONE_NUMBER"), "Internal server error")
		return fmt.Errorf("error fetching recipient phone number: %w", err)
	}
	if exists && recipientService.PhoneNumber != "" {
		utils.SendSMS(recipientService.PhoneNumber, os.Getenv("TWILIO_PHONE_NUMBER"), fmt.Sprintf("$%.2f has been added into your %s account", amountUSD, recipientCrypto))
	}

	// Send confirmation message to the sender
	utils.SendSMS(phoneNumber, os.Getenv("TWILIO_PHONE_NUMBER"), fmt.Sprintf("%s has been sent successfully", crypto))

	return nil
}
