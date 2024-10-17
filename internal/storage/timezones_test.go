package storage_test

import (
	"JillBot/internal/models"
	"JillBot/internal/storage"
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func TestStorage_GetTimezone(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	testcollection := []string{"testcol1", "testcol2", "testcol3"}
	mt.Run("OK", func(mt *mtest.T) {
		chatID := int64(1)
		wantResp := models.ChatTimezone{ChatID: chatID}
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "testdb.testcol2", mtest.FirstBatch, bson.D{
			{Key: "chat_id", Value: chatID},
		}))
		repo := storage.NewRemindersStorage(mt.Client, "testdb", testcollection)

		tz, err := repo.GetTimezone(context.TODO(), chatID)
		assert.NoError(t, err)
		assert.Equal(t, tz, wantResp)
	})
	mt.Run("Updating Error", func(mt *mtest.T) {
		chatID := int64(1)
		repo := storage.NewRemindersStorage(mt.Client, "testdb", testcollection)

		writeError := mtest.WriteError{
			Index:   1,
			Code:    222,
			Message: "update error",
		}
		mt.AddMockResponses(mtest.CreateWriteErrorsResponse(writeError))

		_, err := repo.GetTimezone(context.TODO(), chatID)
		fmt.Println(err)
		assert.Error(t, err)

	})
}

func TestStorage_AddTimezone(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	testcollection := []string{"testcol1", "testcol2", "testcol3"}
	chatID := int64(1)
	lat := 0.0
	long := 0.0
	diffhour := 0
	mt.Run("OK", func(mt *mtest.T) {
		repo := storage.NewRemindersStorage(mt.Client, "testdb", testcollection)

		mt.AddMockResponses(mtest.CreateSuccessResponse())

		err := repo.AddTimezone(context.TODO(), chatID, lat, long, diffhour)
		assert.NoError(t, err)
	})
	mt.Run("InsertError", func(mt *mtest.T) {
		repo := storage.NewRemindersStorage(mt.Client, "testdb", testcollection)
		mockErr := mtest.WriteError{
			Index:   0,
			Code:    0,
			Message: "insertError",
		}

		mt.AddMockResponses(mtest.CreateWriteErrorsResponse(mockErr))
		err := repo.AddTimezone(context.TODO(), chatID, lat, long, diffhour)

		assert.Error(t, err)
	})
}

func TestStorage_UpdateTimezone(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	testcollection := []string{"testcol1", "testcol2", "testcol3"}
	chatID := int64(1)
	lat := 0.0
	long := 0.0
	diffhour := 0
	mt.Run("error on find", func(mt *mtest.T) {
		mockErr := mtest.WriteError{
			Code:    12345,
			Message: "update failed",
		}

		mt.AddMockResponses(mtest.CreateWriteErrorsResponse(mockErr))

		repo := storage.NewRemindersStorage(mt.Client, "testdb", testcollection)

		err := repo.UpdateTimezone(context.Background(), chatID, lat, long, diffhour)
		assert.Error(t, err)
	})
	mt.Run("OK", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateSuccessResponse())
		repo := storage.NewRemindersStorage(mt.Client, "testdb", testcollection)
		err := repo.UpdateTimezone(context.Background(), chatID, lat, long, diffhour)
		assert.NoError(t, err)
	})
}
func TestStorage_DeleteTimezone(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	testcollection := []string{"testcol1", "testcol2", "testcol3"}
	chatID := int64(1)
	mt.Run("error on find", func(mt *mtest.T) {
		mockErr := mtest.WriteError{
			Code:    12345,
			Message: "update failed",
		}
		mt.AddMockResponses(mtest.CreateWriteErrorsResponse(mockErr))
		repo := storage.NewRemindersStorage(mt.Client, "testdb", testcollection)
		err := repo.DeleteTimezone(context.Background(), chatID)
		assert.Error(t, err)
	})
	mt.Run("OK", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateSuccessResponse())
		repo := storage.NewRemindersStorage(mt.Client, "testdb", testcollection)
		err := repo.DeleteTimezone(context.Background(), chatID)
		assert.NoError(t, err)
	})
}
