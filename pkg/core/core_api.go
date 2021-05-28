package core

import (
	"fmt"

	"github.com/quangkeu95/fantom-bot/lib/notification"
)

func (c *Core) AddChatGroup(token string, chatId int64) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	teleBot, err := notification.NewTelegramBot(token, chatId)
	if err != nil {
		return err
	}

	c.socialBots = append(c.socialBots, teleBot)
	return nil
}

func (c *Core) DeleteChatGroup(chatId int64) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var (
		index int
		found bool
	)
	for i, socialBot := range c.socialBots {
		if socialBot.GetChatID() == chatId {
			index = i
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("chat id not found")
	}
	c.socialBots = append(c.socialBots[:index], c.socialBots[index])
	return nil
}
