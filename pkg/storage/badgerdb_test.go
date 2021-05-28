package storage

import (
	"testing"

	"github.com/quangkeu95/fantom-bot/config"
	"github.com/stretchr/testify/suite"
)

type BadgerDBTestSuite struct {
	suite.Suite
	storage *BadgerDB
}

func TestBadgerDBTestSuite(t *testing.T) {
	suite.Run(t, new(BadgerDBTestSuite))
}

func (ts *BadgerDBTestSuite) SetupSuite() {
	config.InitConfig()

	assert := ts.Assert()
	storage, err := NewBadgerDB()
	assert.NoError(err)
	assert.NotNil(storage)
	ts.storage = storage
}

func (ts *BadgerDBTestSuite) TestSet() {
	assert := ts.Assert()
	var (
		key   = "abc123"
		value = float64(101)
	)
	err := ts.storage.Set(key, value)
	assert.NoError(err)

	var result float64
	err = ts.storage.Get(key, &result)
	assert.NoError(err)
	assert.Equal(value, result)
}
