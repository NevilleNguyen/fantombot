package notification

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type TelegramBot struct {
	l      *zap.SugaredLogger
	api    *tgbotapi.BotAPI
	chatId int64
}

func NewTelegramBot() (*TelegramBot, error) {
	l := zap.S()
	token := viper.GetString("telegram.token")
	chatId := viper.GetInt64("telegram.chat_id")

	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		l.Errorw("initialize telegram bot error", "error", err)
		return nil, err
	}
	return &TelegramBot{
		l:      l,
		api:    api,
		chatId: chatId,
	}, nil
}

func (b *TelegramBot) SendMessage(msg string) error {
	sendMsg := tgbotapi.NewMessage(b.chatId, msg)
	_, err := b.api.Send(sendMsg)
	if err != nil {
		b.l.Warnw("telegram bot send message error", "error", err)
		return err
	}
	return err
}

func (b *TelegramBot) SendStructMessage(msg string, keyAndValues map[string]interface{}) error {
	composedMsg := msg
	for key, value := range keyAndValues {
		composedMsg = fmt.Sprintf("%s | %v: %v", composedMsg, key, value)
	}

	return b.SendMessage(composedMsg)
}

func (b *TelegramBot) SendVariadicMessage(msg string, listParams ...interface{}) error {
	if len(listParams)%2 != 0 {
		return fmt.Errorf("list params must be divisible by 2, current length %v", len(listParams))
	}
	composedMsg := msg
	var i int
	for i < len(listParams) {
		key := listParams[i]
		value := listParams[i+1]
		composedMsg = fmt.Sprintf("%s | %v: %v", composedMsg, key, value)
		i += 2
	}
	return b.SendMessage(composedMsg)
}
