package main

import (
	"JillBot/cmd/config"
	"JillBot/cmd/handler"
	"JillBot/internal/service"
	"JillBot/internal/storage"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	config.LoadEnv()
	mongodb := storage.CreateMongoClient(ctx)
	err := mongodb.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatal(err)
	}

	bot, err := telego.NewBot(os.Getenv("BOT_TOKEN"), telego.WithDefaultDebugLogger())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if err != nil {
		log.Fatalf("Failed to set commands: %s", err)
	}
	updates, _ := bot.UpdatesViaLongPolling(nil)
	bh, _ := th.NewBotHandler(bot, updates)
	defer mongodb.Disconnect(ctx)
	defer bh.Stop()
	defer bot.StopLongPolling()

	remindersdb := mongodb.Database("remindersdb")
	store := storage.NewForumStorage(remindersdb, mongodb)
	botSRV := service.NewBotService(store)
	h := handler.NewHandler(bh, botSRV)
	h.InitRoutes()
	go h.StartCheckingReminders(bot)
	bh.Start()
}
