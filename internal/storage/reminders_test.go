package storage_test

import (
	"JillBot/internal/models"
	"JillBot/internal/storage"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func TestStorage_AddReminder(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	testcollection := []string{"testcol1", "testcol2", "testcol3"}
	mt.Run("successful insertion", func(mt *mtest.T) {
		repo := storage.NewRemindersStorage(mt.Client, "testdb", testcollection)
		reminder := models.Reminder{
			IsActive: true,
		}
		mt.AddMockResponses(mtest.CreateSuccessResponse())

		err := repo.AddReminder(context.TODO(), reminder)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})
	mt.Run("InsertError", func(mt *mtest.T) {
		repo := storage.NewRemindersStorage(mt.Client, "testdb", testcollection)
		reminder := models.Reminder{
			IsActive: true,
		}
		mockErr := mtest.WriteError{
			Index:   0,
			Code:    0,
			Message: "insertError",
		}

		mt.AddMockResponses(mtest.CreateWriteErrorsResponse(mockErr))
		err := repo.AddReminder(context.TODO(), reminder)

		assert.Error(t, err)
	})
}

func TestStorage_GetUpcomingReminders(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	testcollection := []string{"testcol1", "testcol2", "testcol3"}
	mt.Run("error on find", func(mt *mtest.T) {
		mockErr := mtest.WriteError{
			Code:    12345,
			Message: "find failed",
		}

		mt.AddMockResponses(mtest.CreateWriteErrorsResponse(mockErr))

		repo := storage.NewRemindersStorage(mt.Client, "testdb", testcollection)

		_, err := repo.GetUpcomingReminders(context.Background())
		if err == nil {
			t.Fatal("expected error, got none")
		}
	})
	mt.Run("OK", func(mt *mtest.T) {
		chatID := int64(12345)
		reminders := []bson.D{
			{{"chat_id", chatID}, {"is_active", true}, {"action", "Reminder 1"}},
			{{"chat_id", chatID}, {"is_active", true}, {"action", "Reminder 2"}},
		}

		mt.AddMockResponses(
			mtest.CreateCursorResponse(1, "test.reminders", mtest.FirstBatch, reminders[0], reminders[1]),
			mtest.CreateCursorResponse(0, "test.reminders", mtest.NextBatch),
		)

		repo := storage.NewRemindersStorage(mt.Client, "testdb", testcollection)

		result,  err := repo.GetUpcomingReminders(context.Background())
		if len(result) != 2 {
			t.Fatalf("expected 2 reminders, got %d", len(result))
		}
		if result[0].Action != "Reminder 1" || result[1].Action != "Reminder 2" {
			t.Fatalf("unexpected reminders: %+v", result)
		}
		assert.NoError(t, err)
	})
	mt.Run("All() error", func(mt *mtest.T) {
		chatID := int64(12345)
		reminders := []bson.D{
			{{"chat_id", chatID}, {"is_active", true}, {"action", "Reminder 1"}},
			{{"chat_id", chatID}, {"is_active", true}, {"action", "Reminder 2"}},
		}

		mockErr := mtest.WriteError{
			Code:    12345,
			Message: "find failed",
		}
		mt.AddMockResponses(
			mtest.CreateCursorResponse(1, "test.reminders", mtest.FirstBatch, reminders[0], reminders[1]),
			mtest.CreateWriteErrorsResponse(mockErr),
		)

		repo := storage.NewRemindersStorage(mt.Client, "testdb", testcollection)

		result,  err := repo.GetUpcomingReminders(context.Background())
		if result != nil{
			t.Fatalf("unexepected result")
		}
		//assert.Equal(t, result, []models.Reminder{})
		assert.Error(t, err)
	})
}

func TestStorage_GetReminders(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	testcollection := []string{"testcol1", "testcol2", "testcol3"}
	mt.Run("error on find", func(mt *mtest.T) {
		mockErr := mtest.WriteError{
			Code:    12345,
			Message: "find failed",
		}
		chatID := int64(1)

		mt.AddMockResponses(mtest.CreateWriteErrorsResponse(mockErr))

		repo := storage.NewRemindersStorage(mt.Client, "testdb", testcollection)

		_, err := repo.GetReminders(context.Background(), chatID)
		if err == nil {
			t.Fatal("expected error, got none")
		}
	})
	mt.Run("OK", func(mt *mtest.T) {
		chatID := int64(12345)
		reminders := []bson.D{
			{{"chat_id", chatID}, {"is_active", true}, {"action", "Reminder 1"}},
			{{"chat_id", chatID}, {"is_active", true}, {"action", "Reminder 2"}},
		}

		mt.AddMockResponses(
			mtest.CreateCursorResponse(1, "test.reminders", mtest.FirstBatch, reminders[0], reminders[1]),
			mtest.CreateCursorResponse(0, "test.reminders", mtest.NextBatch),
		)

		repo := storage.NewRemindersStorage(mt.Client, "testdb", testcollection)

		result,  err := repo.GetReminders(context.Background(), chatID)
		if len(result) != 2 {
			t.Fatalf("expected 2 reminders, got %d", len(result))
		}
		if result[0].Action != "Reminder 1" || result[1].Action != "Reminder 2" {
			t.Fatalf("unexpected reminders: %+v", result)
		}
		assert.NoError(t, err)
	})
	mt.Run("All() error", func(mt *mtest.T) {
		chatID := int64(12345)
		reminders := []bson.D{
			{{"chat_id", chatID}, {"is_active", true}, {"action", "Reminder 1"}},
			{{"chat_id", chatID}, {"is_active", true}, {"action", "Reminder 2"}},
		}

		mockErr := mtest.WriteError{
			Code:    12345,
			Message: "find failed",
		}
		mt.AddMockResponses(
			mtest.CreateCursorResponse(1, "test.reminders", mtest.FirstBatch, reminders[0], reminders[1]),
			mtest.CreateWriteErrorsResponse(mockErr),
		)

		repo := storage.NewRemindersStorage(mt.Client, "testdb", testcollection)

		result,  err := repo.GetReminders(context.Background(), chatID)
		if result != nil{
			t.Fatalf("unexepected result")
		}
		assert.Error(t, err)
	})
	
}

func TestStorage_MarkReminderAsInactive(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	testcollection := []string{"testcol1", "testcol2", "testcol3"}
	mt.Run("error on find", func(mt *mtest.T) {
		mockErr := mtest.WriteError{
			Code:    12345,
			Message: "update failed",
		}
		chatID := int64(1)

		mt.AddMockResponses(mtest.CreateWriteErrorsResponse(mockErr))

		repo := storage.NewRemindersStorage(mt.Client, "testdb", testcollection)

		_, err := repo.GetReminders(context.Background(), chatID)
		assert.Error(t, err)
	})
	mt.Run("OK", func(mt *mtest.T) {
		id := "507f1f77bcf86cd799439011"
		chatID := int64(1)
		mt.AddMockResponses(mtest.CreateSuccessResponse())

		repo := storage.NewRemindersStorage(mt.Client, "testdb", testcollection)

		_, err := repo.MarkReminderAsInactive(context.Background(), chatID, id)
		assert.NoError(t, err)
	})
	mt.Run("InvalidID", func(mt *mtest.T) {
		id := "5d799439011"
		chatID := int64(1)
		mt.AddMockResponses(mtest.CreateSuccessResponse())

		repo := storage.NewRemindersStorage(mt.Client, "testdb", testcollection)

		_, err := repo.MarkReminderAsInactive(context.Background(), chatID, id)
		assert.Error(t, err)
		assert.Equal(t, err, errors.New("invalid ID format"))
	})
}
