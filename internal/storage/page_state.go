package storage

import (
	"JillBot/internal/models"
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (r *RemindersStorage) SetUserPage(ctx context.Context, chatID int64, page int) error {
	filter := bson.M{"chat_id": chatID}
	update := bson.M{
		"$set": bson.M{"page": page},
	}
	opts := options.Update().SetUpsert(true)
	_, err := r.PageState.UpdateOne(ctx, filter, update, opts)
	return err
}

func (r *RemindersStorage) GetUserPage(ctx context.Context, chatID int64) int {
	filter := bson.M{"chat_id": chatID}
	var state models.UserPageState
	err := r.PageState.FindOne(ctx, filter).Decode(&state)
	if err != nil {
		return 0
	}
	return state.Page
}
