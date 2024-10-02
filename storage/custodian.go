package storage

import (
	"context"
	"errors"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Custodian represents a custodian document in the database
type Custodian struct {
	WalletAddress    string             `bson:"wallet_address"`
	Cryptocurrencies map[string]float64 `bson:"cryptocurrencies"`
}

// GetCustodianCollection returns a reference to the custodian collection
func GetCustodianCollection() *mongo.Collection {
	return db.Collection("custodian")
}

// GetCustodianByWalletAddress fetches the custodian data for a given wallet address
func GetCustodianByWalletAddress(ctx context.Context, walletAddress string) (*Custodian, bool, error) {
	collection := GetCustodianCollection()
	var custodian Custodian
	err := collection.FindOne(ctx, bson.M{"wallet_address": walletAddress}).Decode(&custodian)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, false, nil
		}
		log.Printf("Error fetching custodian data: %v", err)
		return nil, false, errors.New("failed to fetch custodian data")
	}
	return &custodian, true, nil
}

// UpdateCustodian updates the custodian data in the database
func UpdateCustodian(ctx context.Context, custodian *Custodian) error {
	collection := GetCustodianCollection()
	filter := bson.M{"wallet_address": custodian.WalletAddress}
	update := bson.M{"$set": custodian}

	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Printf("Error updating custodian data: %v", err)
		return errors.New("failed to update custodian data")
	}
	return nil
}
