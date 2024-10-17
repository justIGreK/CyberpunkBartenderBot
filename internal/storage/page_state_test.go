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

func TestStorage_SetUserPage(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	testcollection := []string{"testcol1", "testcol2", "testcol3"}
	mt.Run("OK", func(mt *mtest.T) {
		chatID := int64(1)
		page := 2
		repo := storage.NewRemindersStorage(mt.Client, "testdb", testcollection)
		mt.AddMockResponses(mtest.CreateSuccessResponse())

		err := repo.SetUserPage(context.TODO(), chatID, page)
		assert.NoError(t, err)
	})
	mt.Run("Updating Error", func(mt *mtest.T) {
		chatID := int64(1)
		page := 2
		repo := storage.NewRemindersStorage(mt.Client, "testdb", testcollection)

		writeError := mtest.WriteError{
			Index:   1,
			Code:    222,
			Message: "update error",
		}
		mt.AddMockResponses(mtest.CreateWriteErrorsResponse(writeError))

		err := repo.SetUserPage(context.TODO(), chatID, page)
		fmt.Println(err)
		assert.Error(t, err)

	})
}

func TestStorage_GetUserPage(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	testcollection := []string{"testcol1", "testcol2", "testcol3"}
	chatID := int64(1)

	mt.Run("OK", func(mt *mtest.T) {

		state := models.UserPageState{Page: 2}
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "testdb.testcol2", mtest.FirstBatch, bson.D{
			{Key: "page", Value: state.Page},
		}))
		repo := storage.NewRemindersStorage(mt.Client, "testdb", testcollection)

		page := repo.GetUserPage(context.TODO(), chatID)
		assert.Equal(t, page, state.Page)
	})
	mt.Run("Getting error", func(mt *mtest.T) {
		wantResp := 0
		mockErr := mtest.WriteError{
			Index:   1,
			Code:    22,
			Message: "fuckingerr",
		}
		mt.AddMockResponses(mtest.CreateWriteErrorsResponse(mockErr))
		repo := storage.NewRemindersStorage(mt.Client, "testdb", testcollection)

		page := repo.GetUserPage(context.TODO(), chatID)
		assert.Equal(t, page, wantResp)
	})
}
