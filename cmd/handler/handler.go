package handler

import (
	"JillBot/internal/models"
	"JillBot/internal/service"
	"context"
	"log"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

type BotSrv interface {
	SetTimezone(ctx context.Context, chatID int64, lat, long float64)
	DeleteTimezone(ctx context.Context, chatID int64) bool
	GetTimezone(ctx context.Context, chatID int64) (models.ChatTimezone, error)
	RemindMe(msg *telego.Message, tz models.ChatTimezone) (string, error)
	GetList(msg *telego.Message) (string, error)
	DeleteReminder(ctx context.Context, msg *telego.Message) (string, error)
	HelpCommand() (string, error)
	GetUpcomingReminders(ctx context.Context) ([]models.Reminder, error)
	MarkReminderAsSent(ctx context.Context, chatID int64, id string) error
	SetUserPage(ctx context.Context, chatID int64, page int) error
	GetUserPage(ctx context.Context, chatID int64) int
	GetListByPage(chatID int64, page int) (string, error)
}
type Handler struct {
	BotHandler *th.BotHandler
	BotSrv
}

func NewHandler(bh *th.BotHandler, botSRV *service.BotSevice) *Handler {

	return &Handler{BotHandler: bh, BotSrv: botSRV}
}

func (h *Handler) InitRoutes() {

	h.BotHandler.Handle(func(bot *telego.Bot, update telego.Update) { // Старт
		chatID := tu.ID(update.Message.Chat.ID)
		response := telego.SendMessageParams{
			ChatID: chatID,
			Text: "Привет, Я Джилл и я призвана помочь тебе с напоминаниями о предстоящих делах\n" +
				"Если ты тут впервые можешь воспользоваться командой /help чтобы узнать о функционале",
		}
		bot.SendMessage(&response)
		requestLocation(bot, chatID)
	}, th.CommandEqual("start"))

	h.BotHandler.Handle(func(bot *telego.Bot, update telego.Update) { // Настройка таймзоны
		chatID := tu.ID(update.Message.Chat.ID)
		h.BotSrv.SetTimezone(context.TODO(), update.Message.Chat.ID, update.Message.Location.Latitude, update.Message.Location.Longitude)
		response := telego.SendMessageParams{
			Text:   "Хорошо, я запомнила",
			ChatID: chatID,
		}
		bot.SendMessage(&response)
	}, func(update telego.Update) bool {
		if update.Message != nil && update.Message.Location != nil {
			return true
		}
		return false
	})

	h.BotHandler.Handle(func(bot *telego.Bot, update telego.Update) { // Удаление часового пояса
		isDelete := h.BotSrv.DeleteTimezone(context.TODO(), update.Message.Chat.ID)
		var text string
		if isDelete {
			text = "Успешно забыто"
		} else {
			text = "Что то пошло не так"
		}
		chatID := tu.ID(update.Message.Chat.ID)
		response := telego.SendMessageParams{
			ChatID: chatID,
			Text:   text,
		}
		bot.SendMessage(&response)
	}, th.CommandEqual("deletelocation"))

	h.BotHandler.Handle(func(bot *telego.Bot, update telego.Update) { // Добавление напоминания
		chatID := tu.ID(update.Message.Chat.ID)
		tz, err := h.BotSrv.GetTimezone(context.TODO(), update.Message.Chat.ID)
		if err != nil {
			response := telego.SendMessageParams{
				ChatID: chatID,
				Text:   "Я не знаю вашего часового пояса. Ты можешь его добавить через /setlocation",
			}
			bot.SendMessage(&response)
			return
		}
		text, err := h.BotSrv.RemindMe(update.Message, tz)
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

	h.BotHandler.Handle(func(bot *telego.Bot, update telego.Update) { // Получение списка напоминаний
		text, err := h.BotSrv.GetList(update.Message)
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

	h.BotHandler.Handle(func(bot *telego.Bot, update telego.Update) { // Удаление напоминания

		text, err := h.BotSrv.DeleteReminder(context.TODO(), update.Message)
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

	h.BotHandler.Handle(func(bot *telego.Bot, update telego.Update) { // Помощь

		text, err := h.BotSrv.HelpCommand()
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

	h.BotHandler.Handle(func(bot *telego.Bot, update telego.Update) { // Установить временную зону
		chatID := tu.ID(update.Message.Chat.ID)
		requestLocation(bot, chatID)
	}, th.CommandEqual("setlocation"))

	h.BotHandler.Handle(func(bot *telego.Bot, update telego.Update) {
		chatID := tu.ID(update.Message.Chat.ID)
		ctx := context.TODO()
		h.BotSrv.SetUserPage(ctx, update.Message.Chat.ID, 0)
		buttons := createPaginationButtons()
		text, err := h.GetListByPage(update.Message.Chat.ID, 0)
		if err != nil {
			log.Printf("\t Не получилось получить список напоминаний, ошибка: %v \n", err)
			text = "Упс, какие то неполадки. Попробуй позже"
		}
		msg := tu.Message(chatID, text).
			WithReplyMarkup(buttons)

		bot.SendMessage(msg)
	}, th.CommandEqual("AYO"))
	h.BotHandler.Handle(func(bot *telego.Bot, update telego.Update) {
		callbackData := update.CallbackQuery.Data
		chat := update.CallbackQuery.Message
		var pageUpdate int
		var text string
		var err error
		if callbackData == "next" {
			pageUpdate = 1
			text, err = h.GetListByPage(update.CallbackQuery.Message.GetChat().ID, pageUpdate)
			if err != nil {
				log.Printf("\t Не получилось получить список напоминаний, ошибка: %v \n", err)
				text = "Упс, какие то неполадки. Попробуй позже"
			}
		} else if callbackData == "refresh" {
			pageUpdate = 0
			text, err = h.GetListByPage(update.CallbackQuery.Message.GetChat().ID, pageUpdate)
			if err != nil {
				log.Printf("\t Не получилось получить список напоминаний, ошибка: %v \n", err)
				text = "Упс, какие то неполадки. Попробуй позже"
			}
		} else if callbackData == "back" {
			pageUpdate = -1
			text, err = h.GetListByPage(update.CallbackQuery.Message.GetChat().ID, pageUpdate)
			if err != nil {
				log.Printf("\t Не получилось получить список напоминаний, ошибка: %v \n", err)
				text = "Упс, какие то неполадки. Попробуй позже"
			}
		}
		buttons := createPaginationButtons()
		bot.EditMessageText(&telego.EditMessageTextParams{
			ChatID:      tu.ID(chat.GetChat().ID),
			MessageID:   chat.GetMessageID(),
			Text:        text,
			ReplyMarkup: buttons,
		})
	}, th.Or(
		th.CallbackDataEqual("back"),
		th.CallbackDataEqual("refresh"),
		th.CallbackDataEqual("next"),
	))
}

func createPaginationButtons() *telego.InlineKeyboardMarkup {

	inlineKeyboard := tu.InlineKeyboard(
		tu.InlineKeyboardRow( // Row 1
			tu.InlineKeyboardButton("Назад").WithCallbackData("back"),
			tu.InlineKeyboardButton("Обновить").WithCallbackData("refresh"),
			tu.InlineKeyboardButton("Вперед").WithCallbackData("next"),
		),
	)
	return inlineKeyboard
}

func requestLocation(bot *telego.Bot, chatID telego.ChatID) {
	locationButton := telego.KeyboardButton{
		Text:            "Поделиться геоданными",
		RequestLocation: true,
	}
	replyMarkup := telego.ReplyKeyboardMarkup{
		Keyboard: [][]telego.KeyboardButton{
			{locationButton},
		},
		ResizeKeyboard:  true,
		OneTimeKeyboard: true,
	}
	response := tu.Message(
		chatID,
		"Для работы мне нужен твой часовой пояс, разреши мне узнать твою геолокацию. Ты всегда можешь удалить эту информацию путем команды /deletelocation",
	).WithReplyMarkup(&replyMarkup)

	bot.SendMessage(response)

}
