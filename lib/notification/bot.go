package notification

type SocialBot interface {
	SendMessage(msg string) error
	SendStructMessage(msg string, params map[string]interface{}) error
	SendVariadicMessage(msg string, keyAndValues ...interface{}) error
}
