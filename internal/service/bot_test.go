package service

import (
	"JillBot/internal/models"
	mock_storage "JillBot/internal/storage/mocks"
	mock_ipgeolocation "JillBot/pkg/ipgeolocation/mocks"
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestService_RemindMe(t *testing.T) {
	type mockBehavior func(r *mock_storage.MockStore, reminder models.Reminder)
	checktime := time.Date(time.Now().UTC().Year(),
		time.Now().UTC().Month(),
		time.Now().UTC().Add(24*time.Hour).Day(), 0, 1, 0, 0, time.UTC)
	id := int64(1)
	testTable := []struct {
		name         string
		msgText      string
		chatID       int64
		reminder     models.Reminder
		OriginalTime time.Time
		UTCtime      time.Time
		action       string
		timezone     models.ChatTimezone
		mockBehavior mockBehavior
		wantErr      bool
		Error        error
		wantResp     string
	}{
		{
			name:    "OK",
			msgText: "/remindme 00:01 test",
			chatID:  int64(1),
			reminder: models.Reminder{
				ChatID: int64(1),
				Action: "test",
				Time: time.Date(time.Now().UTC().Year(),
					time.Now().UTC().Month(),
					time.Now().UTC().Add(24*time.Hour).Day(), 0, 1, 0, 0, time.UTC),
				OriginalTime: time.Date(time.Now().UTC().Year(),
					time.Now().UTC().Month(),
					time.Now().UTC().Add(24*time.Hour).Day(), 0, 1, 0, 0, time.UTC),
			},
			timezone: models.ChatTimezone{id, 0.0, 0.0, 0},
			mockBehavior: func(r *mock_storage.MockStore, reminder models.Reminder) {
				r.EXPECT().AddReminder(gomock.Any(), reminder).Return(nil)
			},
			wantResp: fmt.Sprintf("Напоминание установлено! Дата/время: %v, Действие: test", checktime.Format("2006-01-02 15:04")),
		},
		{
			name:    "OKtimeDate",
			msgText: "/remindme 2040-12-12 12:00 test",
			chatID:  int64(1),
			reminder: models.Reminder{
				ChatID:       int64(1),
				Action:       "test",
				Time:         time.Date(2040, 12, 12, 12, 0, 0, 0, time.UTC),
				OriginalTime: time.Date(2040, 12, 12, 12, 0, 0, 0, time.UTC),
			},
			timezone: models.ChatTimezone{id, 0.0, 0.0, 0},
			mockBehavior: func(r *mock_storage.MockStore, reminder models.Reminder) {
				r.EXPECT().AddReminder(gomock.Any(), reminder).Return(nil)
			},
			wantResp: "Напоминание установлено! Дата/время: 2040-12-12 12:00, Действие: test",
		},
		{
			name:         "ShortMsg",
			msgText:      "/remindme 12:00",
			mockBehavior: func(r *mock_storage.MockStore, reminder models.Reminder) {},
			wantResp: "Пожалуйста укажи дату/время и действие! Например вот так: /remindme 12:00 сходить в магазин\n" +
				"Или например если хочешь на напоминание на завтра или через неделю, укажи точную дату, например /remindme 2024-10-10 12:00 сходить в магазин",
		},
		{
			name:         "InvalidFormat",
			msgText:      "/remindme 1200 test",
			mockBehavior: func(r *mock_storage.MockStore, reminder models.Reminder) {},
			wantErr:      true,
			Error:        errors.New("неправильный формат даты или времени"),
		},
		{
			name:         "InvalidTimeFormat",
			msgText:      "/remindme 25:00 test",
			mockBehavior: func(r *mock_storage.MockStore, reminder models.Reminder) {},
			wantErr:      true,
			Error:        errors.New("ошибка при разборе времени. Формат должен быть HH:mm"),
		},
		{
			name:         "PastTime",
			msgText:      "/remindme 1995-05-25 12:00 test",
			mockBehavior: func(r *mock_storage.MockStore, reminder models.Reminder) {},
			wantErr:      true,
			Error:        errors.New("ошибка: Указанное время уже прошло. Укажите время в будущем"),
		},
	}
	for _, tt := range testTable {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			repo := mock_storage.NewMockStore(ctrl)
			tt.mockBehavior(repo, tt.reminder)
			timeDiffGetter := mock_ipgeolocation.NewMockTimeDiffGetter(ctrl)
			srv := NewBotService(repo, timeDiffGetter)
			msg, err := srv.RemindMe(tt.chatID, tt.msgText, tt.timezone)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, err, tt.Error)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, msg, tt.wantResp)
			}
		})
	}

}

func TestService_timeFormatParse(t *testing.T) {
	type mockBehavior func(r *mock_storage.MockStore, reminder models.Reminder)
	testTable := []struct {
		name         string
		timeSlice    []string
		timezone     models.ChatTimezone
		mockBehavior mockBehavior
		wantErr      bool
		Error        error
		wantResp     ReminderTimes
	}{
		{
			name:      "OK",
			timeSlice: []string{"23:59", ""},
			timezone:  models.ChatTimezone{Diff_hour: 0},
			wantResp: ReminderTimes{
				UTCtime: time.Date(time.Now().UTC().Year(),
					time.Now().UTC().Month(),
					time.Now().UTC().Day(),
					23, 59, 0, 0, time.UTC),
				Originaltime: time.Date(time.Now().UTC().Year(),
					time.Now().UTC().Month(),
					time.Now().UTC().Day(),
					23, 59, 0, 0, time.UTC),
			},
		},
		{
			name:      "UnparseableSlice",
			timeSlice: []string{"1200", ""},
			timezone:  models.ChatTimezone{Diff_hour: 0},
			wantErr:   true,
			Error:     errors.New("ошибка при разборе времени. Формат должен быть HH:mm"),
		},
		{
			name:      "PastButOkay",
			timeSlice: []string{"00:01", ""},
			timezone:  models.ChatTimezone{Diff_hour: 0},
			wantResp: ReminderTimes{
				UTCtime:      time.Date(time.Now().UTC().Year(), time.Now().UTC().Month(), time.Now().UTC().Add(24*time.Hour).Day(), 0, 1, 0, 0, time.UTC),
				Originaltime: time.Date(time.Now().UTC().Year(), time.Now().UTC().Month(), time.Now().UTC().Add(24*time.Hour).Day(), 0, 1, 0, 0, time.UTC),
			},
		},
	}
	for _, tt := range testTable {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			ReminderTimes, err := timeFormatParse(tt.timeSlice, tt.timezone)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, err, tt.Error)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, ReminderTimes, tt.wantResp)
			}
		})
	}

}

func TestService_dateTimeFormatParse(t *testing.T) {
	type mockBehavior func(r *mock_storage.MockStore, reminder models.Reminder)
	testTable := []struct {
		name         string
		timeSlice    []string
		timezone     models.ChatTimezone
		mockBehavior mockBehavior
		wantErr      bool
		Error        error
		wantResp     ReminderTimes
	}{
		{
			name:      "OK",
			timeSlice: []string{"2025-10-16", "12:00"},
			timezone:  models.ChatTimezone{Diff_hour: 0},
			wantResp: ReminderTimes{
				UTCtime:      time.Date(2025, 10, 16, 12, 0, 0, 0, time.UTC),
				Originaltime: time.Date(2025, 10, 16, 12, 0, 0, 0, time.UTC),
			},
		},
		{
			name:      "UnparseableSlice",
			timeSlice: []string{"1200", ""},
			timezone:  models.ChatTimezone{Diff_hour: 0},
			wantErr:   true,
			Error:     errors.New("ошибка при разборе времени. Формат должен быть HH:mm"),
		},
	}
	for _, tt := range testTable {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			ReminderTimes, err := dateTimeFormatParse(tt.timeSlice, tt.timezone)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, err, tt.Error)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, ReminderTimes, tt.wantResp)
			}
		})
	}

}

func TestService_GetListByPage(t *testing.T) {
	type mockBehavior func(r *mock_storage.MockStore, chatID int64, page int)
	testTable := []struct {
		name         string
		chatID       int64
		updatePage   int
		page         int
		mockBehavior mockBehavior
		reminders    []models.Reminder
		wantErr      bool
		Error        error
		wantResp     string
	}{
		{
			name:       "OK",
			chatID:     int64(1),
			updatePage: 0,
			page:       0,
			mockBehavior: func(r *mock_storage.MockStore, chatID int64, page int) {
				r.EXPECT().GetUserPage(context.TODO(), chatID).Return(0)
				r.EXPECT().GetReminders(context.TODO(), chatID).Return(

					[]models.Reminder{
						{
							ID:           "1",
							OriginalTime: time.Date(2025, 10, 16, 12, 0, 0, 0, time.UTC),
							Action:       "test",
						},
						{
							ID:           "2",
							OriginalTime: time.Date(2025, 10, 16, 12, 0, 0, 0, time.UTC),
							Action:       "test",
						},
					}, nil,
				)
				r.EXPECT().SetUserPage(context.TODO(), chatID, 0)
			},
			wantResp: "У вас 2 напоминаний:\n" +
				"ID: 1\n⏰ Время: 2025-10-16 12:00:00\n📋 Действие: test\n\n" +
				"ID: 2\n⏰ Время: 2025-10-16 12:00:00\n📋 Действие: test\n\n" +
				"Страница №1 из 1",
		},
		{
			name:       "OkOutOfRangePage",
			chatID:     int64(1),
			updatePage: 1,
			page:       10000,
			mockBehavior: func(r *mock_storage.MockStore, chatID int64, page int) {
				r.EXPECT().GetUserPage(context.TODO(), chatID).Return(0)
				r.EXPECT().GetReminders(context.TODO(), chatID).Return(

					[]models.Reminder{
						{
							ID:           "1",
							OriginalTime: time.Date(2025, 10, 16, 12, 0, 0, 0, time.UTC),
							Action:       "test",
						},
						{
							ID:           "2",
							OriginalTime: time.Date(2025, 10, 16, 12, 0, 0, 0, time.UTC),
							Action:       "test",
						},
					}, nil,
				)
				r.EXPECT().SetUserPage(context.TODO(), chatID, 0)
			},
			wantResp: "У вас 2 напоминаний:\n" +
				"ID: 1\n⏰ Время: 2025-10-16 12:00:00\n📋 Действие: test\n\n" +
				"ID: 2\n⏰ Время: 2025-10-16 12:00:00\n📋 Действие: test\n\n" +
				"Страница №1 из 1",
		},
		{
			name:       "EmptyList",
			chatID:     int64(1),
			updatePage: 0,
			page:       0,
			mockBehavior: func(r *mock_storage.MockStore, chatID int64, page int) {
				r.EXPECT().GetUserPage(context.TODO(), chatID).Return(0)
				r.EXPECT().GetReminders(context.TODO(), chatID).Return(nil, nil)
			},
			wantResp: "Список напоминаний пуст",
		},
		{
			name:       "GetListError",
			chatID:     int64(1),
			updatePage: 0,
			page:       0,
			mockBehavior: func(r *mock_storage.MockStore, chatID int64, page int) {
				r.EXPECT().GetUserPage(context.TODO(), chatID).Return(0)
				r.EXPECT().GetReminders(context.TODO(), chatID).Return(nil, errors.New("неполадки"))
			},
			wantErr: true,
			Error:   errors.New("неполадки"),
		},
	}
	for _, tt := range testTable {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			repo := mock_storage.NewMockStore(ctrl)
			tt.mockBehavior(repo, tt.chatID, tt.page)
			timeDiffGetter := mock_ipgeolocation.NewMockTimeDiffGetter(ctrl)
			srv := NewBotService(repo, timeDiffGetter)
			msg, err := srv.GetListByPage(tt.chatID, tt.updatePage)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, err, tt.Error)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, msg, tt.wantResp)
			}
		})
	}

}

func TestService_DeleteReminder(t *testing.T) {
	type mockBehavior func(r *mock_storage.MockStore, chatID int64, id string)
	testTable := []struct {
		name         string
		chatID       int64
		msgText      string
		id           string
		mockBehavior mockBehavior
		wantErr      bool
		Error        error
		wantResp     string
	}{
		{
			name:    "OK",
			chatID:  int64(1),
			msgText: "/del 1",
			id:      "1",
			mockBehavior: func(r *mock_storage.MockStore, chatID int64, id string) {
				r.EXPECT().MarkReminderAsInactive(gomock.Any(), chatID, id).Return(int64(1), nil)
			},
			wantResp: "Напоминание удалено успешно",
		},
		{
			name:         "ShortMsg",
			chatID:       int64(1),
			msgText:      "/del 1 1 1 1",
			mockBehavior: func(r *mock_storage.MockStore, chatID int64, id string) {},
			wantResp:     "Пожалуйста укажи айди напоминания! \n Например: /del 6701dca27a3481be8353eee5",
		},
		{
			name:    "DeleteError",
			chatID:  int64(1),
			msgText: "/del 1",
			id:      "1",
			mockBehavior: func(r *mock_storage.MockStore, chatID int64, id string) {
				r.EXPECT().MarkReminderAsInactive(gomock.Any(), chatID, id).Return(int64(1), errors.New("ads"))
			},
			wantErr: true,
			Error:   errors.New("Похоже что-то сломалось..."),
		},
		{
			name:    "NoChanges",
			chatID:  int64(1),
			msgText: "/del 1",
			id:      "1",
			mockBehavior: func(r *mock_storage.MockStore, chatID int64, id string) {
				r.EXPECT().MarkReminderAsInactive(gomock.Any(), chatID, id).Return(int64(0), nil)
			},
			wantResp: "Напоминание не было найдено",
		},
	}
	for _, tt := range testTable {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			repo := mock_storage.NewMockStore(ctrl)
			tt.mockBehavior(repo, tt.chatID, tt.id)
			timeDiffGetter := mock_ipgeolocation.NewMockTimeDiffGetter(ctrl)
			srv := NewBotService(repo, timeDiffGetter)
			msg, err := srv.DeleteReminder(context.TODO(), tt.chatID, tt.msgText)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, err, tt.Error)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, msg, tt.wantResp)
			}
		})
	}

}

func TestService_SetTimezone(t *testing.T) {
	type mockBehavior func(td *mock_ipgeolocation.MockTimeDiffGetter,
		r *mock_storage.MockStore, chatID int64, lat, long float64, diffhour int)
	testTable := []struct {
		name         string
		chatID       int64
		lat          float64
		long         float64
		diffhour     int
		mockBehavior mockBehavior
		wantErr      bool
		Error        error
	}{
		{
			name:   "OKupdate",
			chatID: int64(1),
			lat:    0.0,
			long:   0.0,
			mockBehavior: func(td *mock_ipgeolocation.MockTimeDiffGetter,
				r *mock_storage.MockStore, chatID int64, lat, long float64, diffhour int) {
				td.EXPECT().GetTimeDiff(lat, long).Return(0, nil)
				r.EXPECT().GetTimezone(context.TODO(), chatID).Return(models.ChatTimezone{}, nil)
				r.EXPECT().UpdateTimezone(context.TODO(), chatID, lat, long, diffhour).Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "OKaddNew",
			chatID: int64(1),
			lat:    0.0,
			long:   0.0,
			mockBehavior: func(td *mock_ipgeolocation.MockTimeDiffGetter,
				r *mock_storage.MockStore, chatID int64, lat, long float64, diffhour int) {
				td.EXPECT().GetTimeDiff(lat, long).Return(0, nil)
				r.EXPECT().GetTimezone(context.TODO(), chatID).Return(models.ChatTimezone{}, errors.New("notfound"))
				r.EXPECT().AddTimezone(context.TODO(), chatID, lat, long, diffhour).Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "NoTimediff",
			chatID: int64(1),
			lat:    0.0,
			long:   0.0,
			mockBehavior: func(td *mock_ipgeolocation.MockTimeDiffGetter,
				r *mock_storage.MockStore, chatID int64, lat, long float64, diffhour int) {
				td.EXPECT().GetTimeDiff(lat, long).Return(100, errors.New("error"))
			},
			wantErr: true,
			Error:   errors.New("error"),
		},
		{
			name:   "cantUpdate",
			chatID: int64(1),
			lat:    0.0,
			long:   0.0,
			mockBehavior: func(td *mock_ipgeolocation.MockTimeDiffGetter,
				r *mock_storage.MockStore, chatID int64, lat, long float64, diffhour int) {
				td.EXPECT().GetTimeDiff(lat, long).Return(0, nil)
				r.EXPECT().GetTimezone(context.TODO(), chatID).Return(models.ChatTimezone{}, nil)
				r.EXPECT().UpdateTimezone(context.TODO(), chatID, lat, long, diffhour).Return(errors.New("updating error"))
			},
			wantErr: true,
			Error:   errors.New("updating error"),
		},
		{
			name:   "cantAdd",
			chatID: int64(1),
			lat:    0.0,
			long:   0.0,
			mockBehavior: func(td *mock_ipgeolocation.MockTimeDiffGetter,
				r *mock_storage.MockStore, chatID int64, lat, long float64, diffhour int) {
				td.EXPECT().GetTimeDiff(lat, long).Return(0, nil)
				r.EXPECT().GetTimezone(context.TODO(), chatID).Return(models.ChatTimezone{}, errors.New("notfound"))
				r.EXPECT().AddTimezone(context.TODO(), chatID, lat, long, diffhour).Return(errors.New("adding error"))
			},
			wantErr: true,
			Error:   errors.New("adding error"),
		},
	}
	for _, tt := range testTable {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			repo := mock_storage.NewMockStore(ctrl)
			td := mock_ipgeolocation.NewMockTimeDiffGetter(ctrl)
			tt.mockBehavior(td, repo, tt.chatID, tt.lat, tt.long, tt.diffhour)

			srv := NewBotService(repo, td)
			err := srv.SetTimezone(context.TODO(), tt.chatID, tt.lat, tt.long)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, err, tt.Error)
			} else {
				assert.NoError(t, err)
			}
		})
	}

}

func TestService_GetTimezone(t *testing.T) {
	type mockBehavior func(r *mock_storage.MockStore, chatID int64)
	testTable := []struct {
		name         string
		chatID       int64
		timezone     models.ChatTimezone
		mockBehavior mockBehavior
		wantErr      bool
		Error        error
		wantResp     models.ChatTimezone
	}{
		{
			name:   "OK",
			chatID: int64(1),
			mockBehavior: func(r *mock_storage.MockStore, chatID int64) {
				r.EXPECT().GetTimezone(gomock.Any(), chatID).Return(
					models.ChatTimezone{
						Diff_hour: 0,
					}, nil,
				)
			},
			wantResp: models.ChatTimezone{
				Diff_hour: 0,
			},
		},
		{
			name:   "OK",
			chatID: int64(1),
			mockBehavior: func(r *mock_storage.MockStore, chatID int64) {
				r.EXPECT().GetTimezone(gomock.Any(), chatID).Return(
					models.ChatTimezone{
						Diff_hour: 0,
					}, errors.New("getting error"),
				)
			},
			wantErr: true,
			Error:   errors.New("getting error"),
		},
	}
	for _, tt := range testTable {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			repo := mock_storage.NewMockStore(ctrl)
			td := mock_ipgeolocation.NewMockTimeDiffGetter(ctrl)
			tt.mockBehavior(repo, tt.chatID)

			srv := NewBotService(repo, td)
			tz, err := srv.GetTimezone(context.TODO(), tt.chatID)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, err, tt.Error)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tz, tt.wantResp)
			}
		})
	}

}

func TestService_DeleteTimezone(t *testing.T) {
	type mockBehavior func(r *mock_storage.MockStore, chatID int64)
	testTable := []struct {
		name         string
		chatID       int64
		timezone     models.ChatTimezone
		mockBehavior mockBehavior
		wantErr      bool
		Error        error
		wantResp     bool
	}{
		{
			name:   "OK",
			chatID: int64(1),
			mockBehavior: func(r *mock_storage.MockStore, chatID int64) {
				r.EXPECT().DeleteTimezone(gomock.Any(), chatID).Return(nil)
				r.EXPECT().GetTimezone(gomock.Any(), chatID).Return(models.ChatTimezone{}, errors.New("not found"))
			},
			wantResp: true,
		},
		{
			name:   "DeleteError",
			chatID: int64(1),
			mockBehavior: func(r *mock_storage.MockStore, chatID int64) {
				r.EXPECT().DeleteTimezone(gomock.Any(), chatID).Return(errors.New("not deleted"))
			},
			wantResp: false,
		},
		{
			name:   "ReminderStillExist",
			chatID: int64(1),
			mockBehavior: func(r *mock_storage.MockStore, chatID int64) {
				r.EXPECT().DeleteTimezone(gomock.Any(), chatID).Return(nil)
				r.EXPECT().GetTimezone(gomock.Any(), chatID).Return(models.ChatTimezone{}, nil)
			},
			wantResp: false,
		},
	}
	for _, tt := range testTable {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			repo := mock_storage.NewMockStore(ctrl)
			td := mock_ipgeolocation.NewMockTimeDiffGetter(ctrl)
			tt.mockBehavior(repo, tt.chatID)

			srv := NewBotService(repo, td)
			isDeleted := srv.DeleteTimezone(context.TODO(), tt.chatID)
			assert.Equal(t, isDeleted, tt.wantResp)
		})
	}

}
