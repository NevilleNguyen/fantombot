package core

import (
	"context"
	"testing"

	"github.com/quangkeu95/fantom-bot/config"
	"github.com/stretchr/testify/suite"
)

type GraphqlAltClientTestSuite struct {
	suite.Suite
	client *GraphqlAltClient
}

func TestGraphqlAltClientTestSuite(t *testing.T) {
	suite.Run(t, new(GraphqlAltClientTestSuite))
}

func (ts *GraphqlAltClientTestSuite) SetupSuite() {
	config.InitConfig()

	assert := ts.Assert()
	client, err := NewGraphqlAltClient()
	assert.NoError(err)
	assert.NotNil(client)
	ts.client = client
}

func (ts *GraphqlAltClientTestSuite) TestGetListValidators() {
	assert := ts.Assert()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	validators, err := ts.client.GetListValidators(ctx)
	assert.NoError(err)
	assert.NotNil(validators)
	assert.Greater(len(validators), 0)
}

func (ts *GraphqlAltClientTestSuite) TestGetListFTMTransferByBlock() {
	assert := ts.Assert()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	blockNumber := uint64(4277395)
	logs, err := ts.client.GetListFTMTransferByBlock(ctx, blockNumber)
	assert.NoError(err)
	assert.NotNil(logs)
	assert.Greater(len(logs), 0)
}
