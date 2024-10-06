package handler

import (
	"JillBot/internal/service"
	"context"
	"log"
	"time"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

type Handler struct {
	BotHandler *th.BotHandler
	BotService *service.BotSevice
}

func NewHandler(bh *th.BotHandler, botSRV *service.BotSevice) *Handler {
	return &Handler{BotHandler: bh, BotService: botSRV}
}

func (h *Handler) InitRoutes() {
	h.BotHandler.Handle(func(bot *telego.Bot, update telego.Update) {
		chatID := tu.ID(update.Message.Chat.ID)
		response := telego.SendMessageParams{
			ChatID: chatID,
			Text: "Привет, Я Джилл и я призвана помочь тебе с напоминаниями о предстоящих делах\n" +
				"Если ты тут впервые можешь воспользоваться командой /help чтобы узнать о функционале",
		}
		bot.SendMessage(&response)
	}, th.CommandEqual("start"))
	h.BotHandler.Handle(func(bot *telego.Bot, update telego.Update) {
		text, err := h.BotService.RemindMe(update.Message)
		chatID := tu.ID(update.Message.Chat.ID)
		response := telego.SendMessageParams{
			ChatID: chatID,
		}
		if err != nil {
			response.Text = "Упс, " + err.Error()
		} else {
			response.Text = text
		}
		bot.SendMessage(&response)

	}, th.CommandEqual("remindme"))
	h.BotHandler.Handle(func(bot *telego.Bot, update telego.Update) {
		text, err := h.BotService.GetList(update.Message)
		chatID := tu.ID(update.Message.Chat.ID)
		response := telego.SendMessageParams{
			ChatID: chatID,
		}
		if err != nil {
			response.Text = "Упс, " + err.Error()
		} else {
			response.Text = text
		}
		bot.SendMessage(&response)

	}, th.CommandEqual("list"))
	h.BotHandler.Handle(func(bot *telego.Bot, update telego.Update) {

		text, err := h.BotService.DeleteReminder(context.TODO(), update.Message)
		chatID := tu.ID(update.Message.Chat.ID)
		response := telego.SendMessageParams{
			ChatID: chatID,
		}
		if err != nil {
			response.Text = "Упс, " + err.Error()
		} else {
			response.Text = text
		}
		bot.SendMessage(&response)

	}, th.CommandEqual("del"))
	h.BotHandler.Handle(func(bot *telego.Bot, update telego.Update) {

		text, err := h.BotService.HelpCommand()
		chatID := tu.ID(update.Message.Chat.ID)
		response := telego.SendMessageParams{
			ChatID: chatID,
		}
		if err != nil {
			response.Text = "Упс, " + err.Error()
		} else {
			response.Text = text
		}
		bot.SendMessage(&response)

	}, th.CommandEqual("help"))
}

func (h *Handler) StartCheckingReminders(bot *telego.Bot) {

	for {
		setTimer()
		reminders, err := h.BotService.GetUpcomingReminders(context.TODO())
		if err != nil {
			log.Printf("Ошибка при получении напоминаний: %v", err)
			continue
		}
		time.Sleep(5 * time.Second)

		for _, reminder := range reminders {
			response := telego.SendMessageParams{
				ChatID: tu.ID(reminder.ChatID),
				Text:   reminder.Action,
			}
			_, err := bot.SendMessage(&response)
			if err != nil {
				log.Printf("Ошибка отправки сообщения: %v", err)
				continue
			}
			err = h.BotService.MarkReminderAsSent(context.TODO(), reminder.ChatID, reminder.ID)
			if err != nil {
				log.Printf("Ошибка при обновлении статуса напоминания: %v", err)
			}
		}
	}
}

func setTimer() {
	now := time.Now()
	secondsUntilNextMinute := 60 - now.Second()
	waitTime := time.Duration(secondsUntilNextMinute-5) * time.Second
	time.Sleep(waitTime)
}
