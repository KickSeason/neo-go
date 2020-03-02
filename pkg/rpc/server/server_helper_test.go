package server

import (
	"net/http"
	"os"
	"testing"

	"github.com/CityOfZion/neo-go/config"
	"github.com/CityOfZion/neo-go/pkg/core"
	"github.com/CityOfZion/neo-go/pkg/core/block"
	"github.com/CityOfZion/neo-go/pkg/core/storage"
	"github.com/CityOfZion/neo-go/pkg/core/transaction"
	"github.com/CityOfZion/neo-go/pkg/io"
	"github.com/CityOfZion/neo-go/pkg/network"
	"github.com/CityOfZion/neo-go/pkg/rpc/request"
	"github.com/CityOfZion/neo-go/pkg/util"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

// InvokeFunctionResult struct for testing.
type InvokeFunctionResult struct {
	Script      string              `json:"script"`
	State       string              `json:"state"`
	GasConsumed string              `json:"gas_consumed"`
	Stack       []request.FuncParam `json:"stack"`
	TX          string              `json:"tx,omitempty"`
}

func initServerWithInMemoryChain(t *testing.T) (*core.Blockchain, http.HandlerFunc) {
	var nBlocks uint32

	net := config.ModeUnitTestNet
	configPath := "../../../config"
	cfg, err := config.Load(configPath, net)
	require.NoError(t, err, "could not load config")

	memoryStore := storage.NewMemoryStore()
	logger := zaptest.NewLogger(t)
	chain, err := core.NewBlockchain(memoryStore, cfg.ProtocolConfiguration, logger)
	require.NoError(t, err, "could not create chain")

	go chain.Run()

	// File "./testdata/testblocks.acc" was generated by function core._
	// ("neo-go/pkg/core/helper_test.go").
	// To generate new "./testdata/testblocks.acc", follow the steps:
	// 		1. Rename the function
	// 		2. Add specific test-case into "neo-go/pkg/core/blockchain_test.go"
	// 		3. Run tests with `$ make test`
	f, err := os.Open("testdata/testblocks.acc")
	require.Nil(t, err)
	br := io.NewBinReaderFromIO(f)
	nBlocks = br.ReadU32LE()
	require.Nil(t, br.Err)
	for i := 0; i < int(nBlocks); i++ {
		_ = br.ReadU32LE()
		b := &block.Block{}
		b.DecodeBinary(br)
		require.Nil(t, br.Err)
		require.NoError(t, chain.AddBlock(b))
	}

	serverConfig := network.NewServerConfig(cfg)
	server, err := network.NewServer(serverConfig, chain, logger)
	require.NoError(t, err)
	rpcServer := New(chain, cfg.ApplicationConfiguration.RPC, server, logger)
	handler := http.HandlerFunc(rpcServer.requestHandler)

	return chain, handler
}

type FeerStub struct{}

func (fs *FeerStub) NetworkFee(*transaction.Transaction) util.Fixed8 {
	return 0
}

func (fs *FeerStub) IsLowPriority(util.Fixed8) bool {
	return false
}

func (fs *FeerStub) FeePerByte(*transaction.Transaction) util.Fixed8 {
	return 0
}

func (fs *FeerStub) SystemFee(*transaction.Transaction) util.Fixed8 {
	return 0
}
