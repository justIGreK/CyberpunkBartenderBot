package service

import (
	"JillBot/internal/models"
	"JillBot/internal/storage"
	"JillBot/pkg/ipgeolocation"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
	"time"
)

//go:generate mockgen -source=bot.go -destination=mocks/mock.go

type BotSrv interface {
	SetTimezone(ctx context.Context, chatID int64, lat, long float64) error
	DeleteTimezone(ctx context.Context, chatID int64) bool
	GetTimezone(ctx context.Context, chatID int64) (models.ChatTimezone, error)
	RemindMe(chatID int64, msgText string, tz models.ChatTimezone) (string, error)
	//	GetList(msg *telego.Message) (string, error)
	DeleteReminder(ctx context.Context, chatID int64, msgText string) (string, error)
	HelpCommand() (string, error)
	GetUpcomingReminders(ctx context.Context) ([]models.Reminder, error)
	MarkReminderAsSent(ctx context.Context, chatID int64, id string) error
	SetUserPage(ctx context.Context, chatID int64, page int) error
	GetUserPage(ctx context.Context, chatID int64) int
	GetListByPage(chatID int64, page int) (string, error)
}
type BotSevice struct {
	storage.Store
	ipgeolocation.TimeDiffGetter
}

func NewBotService(store storage.Store, timeDiff ipgeolocation.TimeDiffGetter) *BotSevice {
	return &BotSevice{Store: store,
		TimeDiffGetter: timeDiff}
}

func (b *BotSevice) RemindMe(chatID int64, msgText string, tz models.ChatTimezone) (string, error) {
	log.Println(msgText)
	args := strings.TrimPrefix(msgText, "/remindme")
	args = strings.TrimSpace(args)
	parts := strings.Fields(args)
	if len(parts) < 2 {
		return "Пожалуйста укажи дату/время и действие! Например вот так: /remindme 12:00 сходить в магазин\n" +
			"Или например если хочешь на напоминание на завтра или через неделю, укажи точную дату, например /remindme 2024-10-10 12:00 сходить в магазин", nil

	}
	timeOrDate := parts[:2]
	fmt.Printf("1st part: %s, 2nd part: %s", timeOrDate[0], timeOrDate[1])

	timeFormat := regexp.MustCompile(`^\d{1,2}[:]\d{2}$`)
	dateTimeFormat := regexp.MustCompile(`^\d{4}[-]\d{2}[-]\d{2}`)
	var reminderTime ReminderTimes
	var err error
	var action string
	if timeFormat.MatchString(timeOrDate[0]) {
		reminderTime, err = timeFormatParse(timeOrDate, tz)
		if err != nil {
			return "", err
		}
		action = strings.Join(parts[1:], " ")

	} else if dateTimeFormat.MatchString(timeOrDate[0]) && timeFormat.MatchString(timeOrDate[1]) {
		reminderTime, err = dateTimeFormatParse(timeOrDate, tz)
		if err != nil {
			return "", err
		}
		action = strings.Join(parts[2:], " ")
	} else {
		return "", errors.New("неправильный формат даты или времени")
	}
	if isPastTime(reminderTime.UTCtime) {
		return "", errors.New("ошибка: Указанное время уже прошло. Укажите время в будущем")
	}

	reminder := models.Reminder{
		ChatID:       chatID,
		Action:       action,
		Time:         reminderTime.UTCtime,
		OriginalTime: reminderTime.Originaltime,
	}
	err = b.Store.AddReminder(context.TODO(), reminder)
	if err != nil {
		return "", err
	}
	response := fmt.Sprintf("Напоминание установлено! Дата/время: %s, Действие: %s", reminderTime.Originaltime.Format("2006-01-02 15:04"), action)
	log.Println(response)
	return response, nil
}

type ReminderTimes struct {
	UTCtime      time.Time
	Originaltime time.Time
}

func timeFormatParse(timeSlice []string, tz models.ChatTimezone) (ReminderTimes, error) {
	var responseTimes ReminderTimes
	reminderTime, err := parseTimeWithToday(timeSlice[0])
	if err != nil {
		log.Println(err)
		return responseTimes, errors.New("ошибка при разборе времени. Формат должен быть HH:mm")
	}
	responseTimes.Originaltime = reminderTime
	responseTimes.UTCtime = reminderTime.Add(-time.Duration(tz.Diff_hour) * time.Hour)

	if isPastTime(responseTimes.UTCtime) {
		responseTimes.UTCtime = responseTimes.UTCtime.Add(24 * time.Hour)
		responseTimes.Originaltime = responseTimes.Originaltime.Add(24 * time.Hour)
	}
	return responseTimes, nil
}

func dateTimeFormatParse(timeSlice []string, tz models.ChatTimezone) (ReminderTimes, error) {
	var responseTimes ReminderTimes
	date := strings.Join(timeSlice, " ")
	reminderTime, err := parseTime(date)
	if err != nil {
		log.Println(err)
		return responseTimes, errors.New("ошибка при разборе времени. Формат должен быть HH:mm")
	}
	responseTimes.Originaltime = reminderTime
	responseTimes.UTCtime = reminderTime.Add(-time.Duration(tz.Diff_hour) * time.Hour)
	return responseTimes, nil
}
func (b *BotSevice) GetListByPage(chatID int64, updatePage int) (string, error) {
	ctx := context.TODO()
	page := b.GetUserPage(ctx, chatID)
	page += updatePage

	reminders, err := b.Store.GetReminders(ctx, chatID)
	if err != nil {
		return "", err
	}
	var message string
	if len(reminders) == 0 {
		return "Список напоминаний пуст", nil
	}
	maxPages := len(reminders) / 5
	if page < 0 {
		page = 0
	} else if page > maxPages {
		page = maxPages
	}
	message += fmt.Sprintf("У вас %d напоминаний:\n", len(reminders))
	for i := page * 5; i < len(reminders) && i < page*5+5; i++ {
		message += fmt.Sprintf("ID: %s\n⏰ Время: %s\n📋 Действие: %s\n\n",
			reminders[i].ID, reminders[i].OriginalTime.Format("2006-01-02 15:04:05"), reminders[i].Action)
	}
	message += fmt.Sprintf("Страница №%d из %d", page+1, maxPages+1)
	b.SetUserPage(ctx, chatID, page)
	return message, nil
}

// func (b *BotSevice) GetList(msg *telego.Message) (string, error) {
// 	reminders, err := b.Store.GetReminders(context.TODO(), msg.Chat.ID)
// 	if err != nil {
// 		return "", err
// 	}
// 	var message string
// 	if len(reminders) == 0 {
// 		return "Список напоминаний пуст", nil
// 	}

//		message += fmt.Sprintf("У вас %d напоминаний:\n", len(reminders))
//		for _, reminder := range reminders {
//			message += fmt.Sprintf("ID: %s\n⏰ Время: %s\n📋 Действие: %s\n\n",
//				reminder.ID, reminder.OriginalTime.Format("2006-01-02 15:04:05"), reminder.Action)
//		}
//		return message, nil
//	}
func parseTimeWithToday(timeStr string) (time.Time, error) {
	now := time.Now().UTC()
	layout := "15:04"
	parsedTime, err := time.Parse(layout, timeStr)
	if err != nil {
		return time.Time{}, err
	}
	reminderTime := time.Date(now.Year(), now.Month(), now.Day(), parsedTime.Hour(), parsedTime.Minute(), 0, 0, now.UTC().Location())
	return reminderTime, nil
}
func parseTime(timeStr string) (time.Time, error) {
	now := time.Now().UTC()
	layout := "2006-01-02 15:04"
	parsedTime, err := time.Parse(layout, timeStr)
	if err != nil {
		return time.Time{}, err
	}
	reminderTime := time.Date(parsedTime.Year(), parsedTime.Month(), parsedTime.Day(), parsedTime.Hour(), parsedTime.Minute(), 0, 0, now.Location())
	return reminderTime, nil
}
func isPastTime(date time.Time) bool {
	now := time.Now().UTC()
	fmt.Printf("\n Сейчас: %v \n Проверяемая дата: %v", now, date)
	isPast := date.Before(now)
	return isPast
}

func (s *BotSevice) GetUpcomingReminders(ctx context.Context) ([]models.Reminder, error) {
	return s.Store.GetUpcomingReminders(ctx)
}

func (s *BotSevice) MarkReminderAsSent(ctx context.Context, chatID int64, id string) error {
	_, err := s.Store.MarkReminderAsInactive(ctx, chatID, id)
	return err
}
func (s *BotSevice) DeleteReminder(ctx context.Context, chatID int64, msgText string) (string, error) {
	args := strings.TrimPrefix(msgText, "/del")
	id := strings.TrimSpace(args)
	parts := strings.Fields(args)
	if len(parts) == 0 || len(parts) > 1 {
		return "Пожалуйста укажи айди напоминания! \n Например: /del 6701dca27a3481be8353eee5", nil
	}
	changes, err := s.Store.MarkReminderAsInactive(ctx, chatID, id)
	if err != nil {
		log.Println(err)
		return "", errors.New("Похоже что-то сломалось...")
	}
	if changes == 0 {
		return "Напоминание не было найдено", nil
	}
	return "Напоминание удалено успешно", nil
}
func (s *BotSevice) HelpCommand() (string, error) {
	commands, err := s.LoadCommands()
	if err != nil {
		return "", err
	}
	var helpMessage string
	helpMessage += "Что я могу:\n"
	for _, cmd := range commands {
		helpMessage += fmt.Sprintf("%s - %s\n", cmd.Command, cmd.Description)
	}
	return helpMessage, nil
}
func (s *BotSevice) LoadCommands() ([]models.Command, error) {
	file, err := os.Open("../commands.json")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var commands []models.Command
	byteValue, _ := io.ReadAll(file)
	err = json.Unmarshal(byteValue, &commands)
	if err != nil {
		log.Println(err)
	}
	return commands, nil
}

func (s *BotSevice) SetTimezone(ctx context.Context, chatID int64, lat, long float64) error {
	diffhour, err := s.TimeDiffGetter.GetTimeDiff(lat, long)
	if err != nil {
		log.Println(err)
		return err
	}
	if _, err := s.Store.GetTimezone(ctx, chatID); err == nil {
		err := s.Store.UpdateTimezone(ctx, chatID, lat, long, diffhour)
		if err != nil {
			log.Println(err)
		}
		return err
	}
	err = s.Store.AddTimezone(ctx, chatID, lat, long, diffhour)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil

}

func (s *BotSevice) GetTimezone(ctx context.Context, chatID int64) (models.ChatTimezone, error) {
	tz, err := s.Store.GetTimezone(ctx, chatID)
	if err != nil {
		log.Println(err)
	}
	return tz, err
}

func (s *BotSevice) DeleteTimezone(ctx context.Context, chatID int64) bool {
	err := s.Store.DeleteTimezone(ctx, chatID)
	if err != nil {
		log.Println(err)
		return false
	}
	if _, err := s.Store.GetTimezone(ctx, chatID); err != nil {
		log.Println(err)
		return true
	}
	return false
}
