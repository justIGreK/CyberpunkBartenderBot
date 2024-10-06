package storage

import (
	"JillBot/internal/models"
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type RemindersStorage struct {
	Collection *mongo.Collection
	Client     *mongo.Client
}

func NewForumStorage(db *mongo.Database, client *mongo.Client) *RemindersStorage {
	return &RemindersStorage{
		Collection: db.Collection("reminders"),
		Client:     client,
	}
}

func (r *RemindersStorage) AddReminder(ctx context.Context, reminder models.Reminder) error {
	reminder.IsActive = true
	_, err := r.Collection.InsertOne(ctx, reminder)
	return err
}

func (r *RemindersStorage) GetUpcomingReminders(ctx context.Context) ([]models.Reminder, error) {
	now := time.Now().UTC()
	filter := bson.M{
		"time":      bson.M{"$lte": now.Add(121 * time.Minute)},
		"is_active": true,
	}
	fmt.Println(filter)
	cursor, err := r.Collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	var reminders []models.Reminder
	if err := cursor.All(ctx, &reminders); err != nil {
		return nil, err
	}
	fmt.Println(reminders)
	return reminders, nil
}

func (r *RemindersStorage) MarkReminderAsInactive(ctx context.Context, chatID int64, id string)  (int64, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return 0, errors.New("invalid ID format")
	}
	filter := bson.M{
		"_id": oid,
		"chat_id": chatID,
		"is_active": true,
	}
	update := bson.M{
		"$set": bson.M{
			"is_active": false,
		},
	}
	changes, err := r.Collection.UpdateOne(context.TODO(), filter, update)
	
	return changes.ModifiedCount, err
}

func (r *RemindersStorage) GetReminders(ctx context.Context, chatID int64) ([]models.Reminder, error) {
	filter := bson.M{
		"chat_id": chatID,
		"is_active": true, 
	}

	cursor, err := r.Collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var reminders []models.Reminder
	for cursor.Next(ctx) {
		var reminder models.Reminder
		err := cursor.Decode(&reminder)
		if err != nil {
			return nil, err
		}
		reminders = append(reminders, reminder)
	}

	return reminders, nil
}
