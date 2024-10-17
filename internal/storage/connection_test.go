package storage_test

import (
	"JillBot/internal/models"
	"JillBot/internal/storage"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func TestStorage_CreateMongoClient(t *testing.T) {
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
}
