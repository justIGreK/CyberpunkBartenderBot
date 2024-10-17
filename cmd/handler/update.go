package handler

import (
	"context"
	"log"
	"time"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
)

func (h *Handler) StartCheckingReminders(bot *telego.Bot) {

	for {
		reminders, err := h.BotSrv.GetUpcomingReminders(context.TODO())
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
			err = h.BotSrv.MarkReminderAsSent(context.TODO(), reminder.ChatID, reminder.ID)
			if err != nil {
				log.Printf("Ошибка при обновлении статуса напоминания: %v", err)
			}
		}
		setTimer()
	}
}

func setTimer() {
	now := time.Now()
	secondsUntilNextMinute := 60 - now.Second()
	waitTime := time.Duration(secondsUntilNextMinute-5) * time.Second
	time.Sleep(waitTime)
}
