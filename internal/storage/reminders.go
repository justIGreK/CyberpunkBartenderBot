package storage

import (
	"JillBot/internal/models"
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

//go:generate mockgen -source=reminders.go -destination=mocks/mock.go
type Store interface {
	AddReminder(ctx context.Context, reminder models.Reminder) error
	GetReminders(ctx context.Context, chatID int64) ([]models.Reminder, error)
	GetUpcomingReminders(ctx context.Context) ([]models.Reminder, error)
	MarkReminderAsInactive(ctx context.Context, chatID int64, id string) (int64, error)
	GetTimezone(ctx context.Context, chatID int64) (models.ChatTimezone, error)
	UpdateTimezone(ctx context.Context, chatID int64, lat, long float64, diffhour int) error
	AddTimezone(ctx context.Context, chatID int64, lat, long float64, diffhour int) error
	DeleteTimezone(ctx context.Context, chatID int64) error
	SetUserPage(ctx context.Context, chatID int64, page int) error
	GetUserPage(ctx context.Context, chatID int64) int
}

type RemindersStorage struct {
	Reminders     *mongo.Collection
	ChatTimezones *mongo.Collection
	PageState     *mongo.Collection
}

func NewRemindersStorage(client *mongo.Client, dbname string, collectionnames []string) *RemindersStorage {
	return &RemindersStorage{
		Reminders:     client.Database(dbname).Collection(collectionnames[0]),
		ChatTimezones: client.Database(dbname).Collection(collectionnames[1]),
		PageState:     client.Database(dbname).Collection(collectionnames[2]),
	}
}

func (r *RemindersStorage) AddReminder(ctx context.Context, reminder models.Reminder) error {
	reminder.IsActive = true
	_, err := r.Reminders.InsertOne(ctx, reminder)
	return err
}
func (r *RemindersStorage) GetUpcomingReminders(ctx context.Context) ([]models.Reminder, error) {
	now := time.Now().UTC()
	filter := bson.M{
		"time":      bson.M{"$lte": now.Add(60 * time.Second)},
		"is_active": true,
	}
	fmt.Println(filter)
	cursor, err := r.Reminders.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var reminders []models.Reminder
	if err := cursor.All(ctx, &reminders); err != nil {
		return nil, err
	}
	fmt.Println(reminders)
	return reminders, nil
}

func (r *RemindersStorage) MarkReminderAsInactive(ctx context.Context, chatID int64, id string) (int64, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return 0, errors.New("invalid ID format")
	}
	filter := bson.M{
		"_id":       oid,
		"chat_id":   chatID,
		"is_active": true,
	}
	update := bson.M{
		"$set": bson.M{
			"is_active": false,
		},
	}
	changes, err := r.Reminders.UpdateOne(context.TODO(), filter, update)

	return changes.ModifiedCount, err
}

func (r *RemindersStorage) GetReminders(ctx context.Context, chatID int64) ([]models.Reminder, error) {
	filter := bson.M{
		"chat_id":   chatID,
		"is_active": true,
	}
	cursor, err := r.Reminders.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var reminders []models.Reminder
	err = cursor.All(ctx, &reminders)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return reminders, nil
}
