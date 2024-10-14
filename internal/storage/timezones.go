package storage

import (
	"JillBot/internal/models"
	"context"

	"go.mongodb.org/mongo-driver/bson"
)

func (r *RemindersStorage) GetTimezone(ctx context.Context, chatID int64) (models.ChatTimezone, error) {
	var tz models.ChatTimezone
	err := r.ChatTimezones.FindOne(ctx, bson.M{"chat_id": chatID}).Decode(&tz)
	if err != nil {
		return tz, err
	}
	return tz, nil
}
func (r *RemindersStorage) AddTimezone(ctx context.Context, chatID int64, lat, long float64) error {
	tz := models.ChatTimezone{
		ChatID:   chatID,
		Latitude: lat,
		Longitude: long,
	}
	_, err := r.ChatTimezones.InsertOne(ctx, tz)
	return err
}
func (r *RemindersStorage) UpdateTimezone(ctx context.Context, chatID int64, lat, long float64) error {
	updateTZ := bson.M{
		"$set": bson.M{
			"lat": lat,
			"long": long,
		},
	}
	filter := bson.M{"chat_id": chatID}
	_, err := r.ChatTimezones.UpdateOne(context.TODO(), filter, updateTZ)
	if err != nil {
		return err
	}
	return nil
}

func (r *RemindersStorage) DeleteTimezone(ctx context.Context, chatID int64) error {
	filter := bson.M{"chat_id": chatID}
	_, err := r.ChatTimezones.DeleteOne(ctx, filter)
	return err
}
