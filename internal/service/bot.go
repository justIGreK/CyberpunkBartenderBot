package service

import (
	"JillBot/internal/models"
	"JillBot/internal/storage"
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

	"github.com/mymmrac/telego"
)

type BotSevice struct {
	Store *storage.RemindersStorage
}

func NewBotService(store *storage.RemindersStorage) *BotSevice {
	return &BotSevice{Store: store}
}

func (b *BotSevice) RemindMe(msg *telego.Message) (string, error) {

	args := strings.TrimPrefix(msg.Text, "/remindme")
	args = strings.TrimSpace(args)
	parts := strings.Fields(args)
	if len(parts) < 2 {
		return "Пожалуйста укажи дату/время и действие! Например вот так: /remindme 12:00 сходить в магазин", nil

	}
	timeOrDate := parts[:2]
	fmt.Printf("1st part: %s, 2nd part: %s", timeOrDate[0], timeOrDate[1])
	action := strings.Join(parts[1:], " ")

	timeFormat := regexp.MustCompile(`^\d{1,2}[:]\d{2}$`)
	dateTimeFormat := regexp.MustCompile(`^\d{4}[-]\d{2}[-]\d{2}`)
	var reminderTime time.Time
	var err error

	if timeFormat.MatchString(timeOrDate[0]) {
		reminderTime, err = parseTimeWithToday(timeOrDate[0])
		if err != nil {
			log.Println(err)
			return "", fmt.Errorf("ошибка при разборе времени. Формат должен быть HH:mm.")
		}

	} else if dateTimeFormat.MatchString(timeOrDate[0]) {
		date := strings.Join(timeOrDate, " ")
		fmt.Println(date)
		reminderTime, err = time.Parse("2006-01-02 15:04", date)
		if err != nil {
			log.Println(err)
			return "", errors.New("ошибка при разборе даты. Формат должен быть YYYY-MM-DD HH:mm")
		}
		action = strings.Join(parts[2:], " ")
	} else {
		return "", errors.New("Неправильный формат даты или времени")
	}
	if isPastTime(reminderTime) {
		return "", errors.New("Ошибка: Указанное время уже прошло. Укажите время в будущем")
	}

	reminder := models.Reminder{
		ChatID: msg.Chat.ID,
		Action: action,
		Time:   reminderTime,
	}
	b.Store.AddReminder(context.TODO(), reminder)
	response := fmt.Sprintf("Напоминание установлено! Дата/время: %s, Действие: %s", reminderTime.Format("2006-01-02 15:04"), action)
	return response, nil
}

func (b *BotSevice) GetList(msg *telego.Message) (string, error) {
	reminders, err := b.Store.GetReminders(context.TODO(), msg.Chat.ID)
	if err != nil {
		return "", err
	}
	var message string
	if len(reminders) == 0 {
		return "Список напоминаний пуст", nil
	}
	message += fmt.Sprintf("У вас %d напоминаний:\n", len(reminders))
	for _, reminder := range reminders {
		message += fmt.Sprintf("ID: %s\n⏰ Время: %s\n📋 Действие: %s\n\n",
			reminder.ID, reminder.Time.Format("2006-01-02 15:04:05"), reminder.Action)
	}
	return message, nil
}
func parseTimeWithToday(timeStr string) (time.Time, error) {
	now := time.Now()
	layout := "15:04"
	parsedTime, err := time.Parse(layout, timeStr)
	if err != nil {
		return time.Time{}, err
	}
	reminderTime := time.Date(now.Year(), now.Month(), now.Day(), parsedTime.Hour(), parsedTime.Minute(), 0, 0, now.Location())
	return reminderTime, nil
}

func isPastTime(date time.Time) bool {
	now := time.Now()
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

func (s *BotSevice) DeleteReminder(ctx context.Context, msg *telego.Message) (string, error) {
	args := strings.TrimPrefix(msg.Text, "/del")
	id := strings.TrimSpace(args)
	parts := strings.Fields(args)
	if len(parts) == 0 || len(parts) > 1 {
		return "Пожалуйста укажи айди напоминания! \n Например: /del 6701dca27a3481be8353eee5", nil
	}
	changes, err := s.Store.MarkReminderAsInactive(ctx, msg.Chat.ID, id)
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
