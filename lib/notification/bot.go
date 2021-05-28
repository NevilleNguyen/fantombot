package notification

type SocialBot interface {
	SendMessage(msg string) error
	GetChatID() int64
}
