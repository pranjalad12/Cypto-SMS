package storage

import (
	"context"
	"errors"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// SmsService represents an SMS service document in the database
type SmsService struct {
	WalletAddress string  `bson:"wallet_address"`
	PhoneNumber   string  `bson:"phone_number"`
	Passkey       string  `bson:"passkey"`
	Limit         float64 `bson:"limit"`
	PublicKey     string  `bson:"public_key"`
}

// GetSmsServiceCollection returns a reference to the sms_service collection
func GetSmsServiceCollection() *mongo.Collection {
	return db.Collection("sms_service")
}

// CheckWalletExistsInSmsService checks if a wallet address exists in the sms_service collection
func CheckWalletExistsInSmsService(ctx context.Context, walletAddress string) (*SmsService, bool, error) {
	collection := GetSmsServiceCollection()
	var service SmsService
	err := collection.FindOne(ctx, bson.M{"wallet_address": walletAddress}).Decode(&service)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, false, nil
		}
		log.Printf("Error checking wallet address: %v", err)
		return nil, false, errors.New("failed to check wallet address")
	}
	return &service, true, nil
}

// CheckPhoneNumberExistsInSmsService checks if a phone number is used by any wallet address
func CheckPhoneNumberExistsInSmsService(ctx context.Context, phoneNumber string) (*SmsService, bool, error) {
	collection := GetSmsServiceCollection()
	var service SmsService
	err := collection.FindOne(ctx, bson.M{"phone_number": phoneNumber}).Decode(&service)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, false, nil
		}
		log.Printf("Error checking phone number: %v", err)
		return nil, false, errors.New("failed to check phone number")
	}
	return &service, true, nil
}

// CreateSmsService creates a new SMS service document in the sms_service collection
func CreateSmsService(ctx context.Context, service SmsService) error {
	collection := GetSmsServiceCollection()
	_, err := collection.InsertOne(ctx, service)
	if err != nil {
		log.Printf("Error adding SMS service: %v", err)
		return errors.New("failed to add SMS service")
	}
	return nil
}

// UpdateSmsService updates an existing SMS service document in the sms_service collection
func UpdateSmsService(ctx context.Context, walletAddress string, passkey string, limit float64) error {
	collection := GetSmsServiceCollection()
	filter := bson.M{"wallet_address": walletAddress}
	update := bson.M{"$set": bson.M{"passkey": passkey, "limit": limit}}
	options := options.Update().SetUpsert(true)

	_, err := collection.UpdateOne(ctx, filter, update, options)
	if err != nil {
		log.Printf("Error updating SMS service: %v", err)
		return errors.New("failed to update SMS service")
	}
	return nil
}

// UpdatePhoneNumber updates the phone number for a given wallet address
func UpdatePhoneNumber(ctx context.Context, walletAddress string, phoneNumber string) error {
	collection := GetSmsServiceCollection()

	// Check if the phone number is already used by another wallet address
	existingService, exists, err := CheckPhoneNumberExistsInSmsService(ctx, phoneNumber)
	if err != nil {
		return err
	}
	if exists {
		// If the phone number is used by another wallet address, make it empty in the existing record
		filter := bson.M{"wallet_address": existingService.WalletAddress}
		update := bson.M{"$set": bson.M{"phone_number": ""}}
		_, err := collection.UpdateOne(ctx, filter, update)
		if err != nil {
			log.Printf("Error updating existing phone number: %v", err)
			return errors.New("failed to update existing phone number")
		}
	}

	// Update the provided wallet address with the phone number
	filter := bson.M{"wallet_address": walletAddress}
	update := bson.M{"$set": bson.M{"phone_number": phoneNumber}}
	_, err = collection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Printf("Error updating phone number for wallet address: %v", err)
		return errors.New("failed to update phone number for wallet address")
	}

	return nil
}
func ListSmsServices(ctx context.Context) ([]SmsService, error) {
	collection := GetSmsServiceCollection()
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var services []SmsService
	if err := cursor.All(ctx, &services); err != nil {
		return nil, err
	}
	return services, nil
}
