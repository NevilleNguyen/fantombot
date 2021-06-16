package config

import (
	"strings"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const (
	ContactBookFlag   = "contact_book"
	ValidatorBookFlag = "validator_book"
)

type Contact struct {
	Address string
	Name    string
}

func GetContact() map[string]string {
	l := zap.S()
	var result = make(map[string]string)

	var resp = make([]Contact, 0)
	if err := viper.UnmarshalKey(ContactBookFlag, &resp); err != nil {
		l.Errorw("error parse contact book", "error", err)
		return result
	}

	for _, contact := range resp {
		result[strings.ToLower(contact.Address)] = contact.Name
	}

	return result
}

type ValidatorContact struct {
	ID   uint64
	Name string
}

func GetValidatorContact() map[uint64]string {
	l := zap.S()
	var result = make(map[uint64]string)

	var resp = make([]ValidatorContact, 0)
	if err := viper.UnmarshalKey(ValidatorBookFlag, &resp); err != nil {
		l.Errorw("error parse validator book", "error", err)
		return result
	}

	for _, validator := range resp {
		result[validator.ID] = validator.Name
	}
	return result
}
