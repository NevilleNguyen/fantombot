package core

import (
	"context"
	"log"
	"testing"

	"github.com/quangkeu95/fantom-bot/config"
	"github.com/stretchr/testify/suite"
)

type SFCCLientTestSuite struct {
	suite.Suite
	client *SFCClient
}

func TestSFCCLientTestSuite(t *testing.T) {
	suite.Run(t, new(SFCCLientTestSuite))
}

func (ts *SFCCLientTestSuite) SetupSuite() {
	config.InitConfig()

	assert := ts.Assert()
	nodeClient, err := NewNodeClient()
	assert.NoError(err)
	assert.NotNil(nodeClient)
	wsClient, err := NewWsClient()
	assert.NoError(err)
	assert.NotNil(wsClient)
	client, err := NewSFCClient(nodeClient, wsClient)
	assert.NoError(err)
	assert.NotNil(client)

	ts.client = client
}

func (ts *SFCCLientTestSuite) TestGetCreatedValidatorByID() {
	assert := ts.Assert()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var (
		ids = []uint64{1}
	)
	validators, err := ts.client.GetCreatedValidatorByID(ctx, ids)
	assert.NoError(err)
	assert.NotNil(validators)
	assert.Equal(10, len(validators))
}

func (ts *SFCCLientTestSuite) TestGetValidatorByID() {
	assert := ts.Assert()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	validator, err := ts.client.GetValidatorByID(ctx, uint64(1))
	assert.NoError(err)
	assert.Equal(uint64(1), validator.ID)
	assert.Equal(true, validator.IsActive)

	validator, err = ts.client.GetValidatorByID(ctx, uint64(60))
	assert.NoError(err)
	assert.Equal(uint64(60), validator.ID)
	assert.Equal(true, validator.IsActive)

	validator, err = ts.client.GetValidatorByID(ctx, uint64(4))
	assert.NoError(err)
	assert.Equal(uint64(4), validator.ID)
	assert.Equal(false, validator.IsActive)
}

func (ts *SFCCLientTestSuite) TestGetDelegate() {
	assert := ts.Assert()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	toBlock := uint64(75147450)
	delegateInfo, err := ts.client.GetDelegateInfoByBlock(ctx, 7514749, &toBlock)
	assert.NoError(err)
	assert.NotNil(delegateInfo)
	assert.Greater(len(delegateInfo), 0)
	lastDelegate := delegateInfo[len(delegateInfo)-1]

	log.Println(lastDelegate)
}

func (ts *SFCCLientTestSuite) TestGetTxByHash() {
	assert := ts.Assert()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := ts.client.GetTxByHash(ctx, "0x217c056f854f99842fb40ace08f3d57f9cc78379bac0b09c6850cce742ae4c92")
	assert.NoError(err)
}
