package fetcher

import (
	"context"
	"testing"

	"github.com/quangkeu95/fantom-bot/config"
	"github.com/stretchr/testify/suite"
)

type GraphqlClientTestSuite struct {
	suite.Suite
	client *GraphqlClient
}

func TestGraphqlClientTestSuite(t *testing.T) {
	suite.Run(t, new(GraphqlClientTestSuite))
}

func (ts *GraphqlClientTestSuite) SetupSuite() {
	config.InitConfig()

	assert := ts.Assert()
	client, err := NewGraphqlClient()
	assert.NoError(err)
	assert.NotNil(client)
	ts.client = client
}

func (ts *GraphqlClientTestSuite) TestGetListValidators() {
	assert := ts.Assert()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	validators, err := ts.client.GetListValidators(ctx)
	assert.NoError(err)
	assert.NotNil(validators)
	assert.Greater(len(validators), 0)
}

func (ts *GraphqlClientTestSuite) TestGetListFTMTransferByBlock() {
	assert := ts.Assert()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	blockNumber := uint64(4277395)
	logs, err := ts.client.GetListFTMTransferByBlock(ctx, blockNumber)
	assert.NoError(err)
	assert.NotNil(logs)
	assert.Equal(2, len(logs))
}
