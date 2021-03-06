package notification

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"go.uber.org/zap"
)

const (
	EmojiWhale     = "\U0001F40B"
	EmojiCheckMark = "\U00002705"
	EmojiCrossMark = "\U0000274C"
	EmojiStar      = "\U00002B50"
	EmojiLock      = "\U0001F512"
	EmojiUnlock    = "\U0001F513"
)

type TelegramBot struct {
	l      *zap.SugaredLogger
	api    *tgbotapi.BotAPI
	chatId int64
}

func NewTelegramBot(token string, chatId int64) (*TelegramBot, error) {
	l := zap.S()

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
	sendMsg.ParseMode = tgbotapi.ModeHTML
	sendMsg.DisableWebPagePreview = true

	_, err := b.api.Send(sendMsg)
	if err != nil {
		b.l.Warnw("telegram bot send message error", "error", err)
		return err
	}
	return err
}

func (b *TelegramBot) GetChatID() int64 {
	return b.chatId
}
