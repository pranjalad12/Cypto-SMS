package storage

import (
	"context"
	"errors"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TwoFactorAuth represents a 2FA document in the database
type TwoFactorAuth struct {
	PhoneNumber string `bson:"phone_number"`
	Code        string `bson:"code"`
}

// GetTwoFactorAuthCollection returns a reference to the 2fa collection
func GetTwoFactorAuthCollection() *mongo.Collection {
	return db.Collection("2fa")
}

// Generate2FACodeAndStore generates a 2FA code and stores it in the 2fa collection// Generate2FACodeAndStore generates a 2FA code and stores it in the 2fa collection
func Generate2FACodeAndStore(ctx context.Context, phoneNumber string, code string) error {
	collection := GetTwoFactorAuthCollection()
	filter := bson.M{"phone_number": phoneNumber}
	update := bson.M{"$set": bson.M{"code": code}}

	options := options.Update().SetUpsert(true) // Directly pass the boolean value

	_, err := collection.UpdateOne(ctx, filter, update, options)
	if err != nil {
		log.Printf("Error updating 2FA code: %v", err)
		return errors.New("failed to update 2FA code")
	}
	return nil
}
// Verify2FACode verifies the 2FA code for a given phone number
func Verify2FACode(ctx context.Context, phoneNumber string, code string) (bool, error) {
	collection := GetTwoFactorAuthCollection()
	var twoFactorAuth TwoFactorAuth
	err := collection.FindOne(ctx, bson.M{"phone_number": phoneNumber}).Decode(&twoFactorAuth)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return false, nil
		}
		log.Printf("Error checking 2FA code: %v", err)
		return false, errors.New("failed to check 2FA code")
	}
	return twoFactorAuth.Code == code, nil
}
func Store2FACode(ctx context.Context, phoneNumber, code string) error {
	collection := GetTwoFactorAuthCollection()
	_, err := collection.UpdateOne(
		ctx,
		bson.M{"phone_number": phoneNumber},
		bson.M{"$set": bson.M{"2fa_code": code}},
		options.Update().SetUpsert(true), // Pass the boolean directly
	)
	if err != nil {
		log.Printf("Failed to store 2FA code: %v", err)
		return err
	}
	return nil
}
