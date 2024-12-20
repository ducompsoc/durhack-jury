package database

import (
	"context"
	"errors"
	"server/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

// GetOptions gets the options from the database
func GetOptions(db *mongo.Database) (*models.Options, error) {
	var options models.Options
	err := db.Collection("options").FindOne(context.Background(), gin.H{}).Decode(&options)

	// If options does not exist, create it
	if errors.Is(err, mongo.ErrNoDocuments) {
		options = *models.NewOptions()
		_, err = db.Collection("options").InsertOne(context.Background(), options)
		return &options, err
	}

	return &options, err
}

// UpdateOptions updates the clock in the database
func UpdateClock(db *mongo.Database, clock *models.ClockState) error {
	_, err := db.Collection("options").UpdateOne(context.Background(), gin.H{}, gin.H{"$set": gin.H{"clock": clock}})
	return err
}

// GetCategories gets the categories from the database
func GetCategories(db *mongo.Database) ([]string, error) {
	var options models.Options
	err := db.Collection("options").FindOne(context.Background(), gin.H{}).Decode(&options)
	return options.Categories, err
}

// GetMinViews gets the minimum views option from the database
func GetMinViews(db *mongo.Database) (int64, error) {
	var options models.Options
	err := db.Collection("options").FindOne(context.Background(), gin.H{}).Decode(&options)
	return options.MinViews, err
}

// GetBatchRankingSize gets the ranking batch size option from the database
func GetBatchRankingSize(db *mongo.Database) (int64, error) {
	var options models.Options
	err := db.Collection("options").FindOne(context.Background(), gin.H{}).Decode(&options)
	return options.BatchRankingSize, err
}

// GetJudgingEnded gets the judgingEnded flag from the database
func GetJudgingEnded(db *mongo.Database) (bool, error) {
	options, err := GetOptions(db)
	if options != nil {
		return options.JudgingEnded, err
	}
	return false, err
}
