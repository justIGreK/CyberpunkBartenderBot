package service

import (
	"context"
	"log"
)

func (s *BotSevice) SetUserPage(ctx context.Context, chatID int64, page int) error {
	err := s.Store.SetUserPage(ctx, chatID, page)
	if err != nil {
		log.Println(err)
	}
	return err
}

func (s *BotSevice) GetUserPage(ctx context.Context, chatID int64) int {
	page := s.Store.GetUserPage(ctx, chatID)
	return page
}
