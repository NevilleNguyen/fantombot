package fetcher

import (
	"context"
	"testing"

	"github.com/quangkeu95/fantom-bot/config"
	"github.com/stretchr/testify/suite"
)

type NodeClientTestSuite struct {
	suite.Suite
	client *NodeClient
}

func TestNodeClientTestSuite(t *testing.T) {
	suite.Run(t, new(NodeClientTestSuite))
}

func (ts *NodeClientTestSuite) SetupSuite() {
	config.InitConfig()

	assert := ts.Assert()
	client, err := NewNodeClient()
	assert.NoError(err)
	assert.NotNil(client)
	ts.client = client
}

func (ts *NodeClientTestSuite) TestGetListFTMTransferByBlock() {
	assert := ts.Assert()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	blockNumber := uint64(4277395)
	logs, err := ts.client.GetListFTMTransferByBlock(ctx, blockNumber)
	assert.NoError(err)
	assert.NotNil(logs)
	assert.Equal(2, len(logs))
}
