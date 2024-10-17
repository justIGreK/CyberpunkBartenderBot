package handler

import (
	"JillBot/internal/service"
	"context"
	"log"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

//go:generate mockgen -source=handler.go -destination=mocks/mock.go

type BotHandler interface {
	Handle(handler th.Handler, predicates ...th.Predicate)
}
type Handler struct {
	BotHandler
	service.BotSrv
}

func NewHandler(bh *th.BotHandler, botSRV service.BotSrv) *Handler {

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
		err := h.BotSrv.SetTimezone(context.TODO(), update.Message.Chat.ID, update.Message.Location.Latitude, update.Message.Location.Longitude)
		var text string
		if err != nil {
			text = "Упс, что-то пошло не так"
		} else {
			text = "Хорошо, я запомнила"
		}
		response := telego.SendMessageParams{
			Text:   text,
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
		isDeleted := h.BotSrv.DeleteTimezone(context.TODO(), update.Message.Chat.ID)
		var text string
		if isDeleted {
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
		text, err := h.BotSrv.RemindMe(update.Message.Chat.ID, update.Message.Text, tz)
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

	// h.BotHandler.Handle(func(bot *telego.Bot, update telego.Update) { // Получение списка напоминаний
	// 	text, err := h.BotSrv.GetList(update.Message)
	// 	chatID := tu.ID(update.Message.Chat.ID)
	// 	response := telego.SendMessageParams{
	// 		ChatID: chatID,
	// 	}
	// 	if err != nil {
	// 		response.Text = "Упс, " + err.Error()
	// 	} else {
	// 		response.Text = text
	// 	}
	// 	bot.SendMessage(&response)

	// }, th.CommandEqual("list"))

	h.BotHandler.Handle(func(bot *telego.Bot, update telego.Update) { // Удаление напоминания

		text, err := h.BotSrv.DeleteReminder(context.TODO(), update.Message.Chat.ID, update.Message.Text)
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
	}, th.CommandEqual("list"))
	h.BotHandler.Handle(func(bot *telego.Bot, update telego.Update) {
		// txt := update.EditedMessage
		// fmt.Println(txt)
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
