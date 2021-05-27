package notification

type SocialBot interface {
	SendMessage(msg string) error
}
