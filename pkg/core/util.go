package core

import (
	"time"

	"github.com/nspcc-dev/neo-go/pkg/config"
	"github.com/nspcc-dev/neo-go/pkg/core/block"
	"github.com/nspcc-dev/neo-go/pkg/core/transaction"
	"github.com/nspcc-dev/neo-go/pkg/crypto/hash"
	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neo-go/pkg/io"
	"github.com/nspcc-dev/neo-go/pkg/smartcontract"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/nspcc-dev/neo-go/pkg/vm"
	"github.com/nspcc-dev/neo-go/pkg/vm/emit"
	"github.com/nspcc-dev/neo-go/pkg/vm/opcode"
)

var (
	// governingTokenTX represents transaction that is used to create
	// governing (NEO) token. It's a part of the genesis block.
	governingTokenTX transaction.Transaction

	// utilityTokenTX represents transaction that is used to create
	// utility (GAS) token. It's a part of the genesis block. It's mostly
	// useful for its hash that represents GAS asset ID.
	utilityTokenTX transaction.Transaction

	// ecdsaVerifyInteropPrice returns the price of Neo.Crypto.ECDsaVerify
	// syscall to calculate NetworkFee for transaction
	ecdsaVerifyInteropPrice = util.Fixed8(100000)
)

// createGenesisBlock creates a genesis block based on the given configuration.
func createGenesisBlock(cfg config.ProtocolConfiguration) (*block.Block, error) {
	validators, err := getValidators(cfg)
	if err != nil {
		return nil, err
	}

	nextConsensus, err := getNextConsensusAddress(validators)
	if err != nil {
		return nil, err
	}

	base := block.Base{
		Version:       0,
		PrevHash:      util.Uint256{},
		Timestamp:     uint64(time.Date(2016, 7, 15, 15, 8, 21, 0, time.UTC).Unix()) * 1000, // Milliseconds.
		Index:         0,
		NextConsensus: nextConsensus,
		Script: transaction.Witness{
			InvocationScript:   []byte{},
			VerificationScript: []byte{byte(opcode.PUSH1)},
		},
	}

	b := &block.Block{
		Base: base,
		Transactions: []*transaction.Transaction{
			deployNativeContracts(),
		},
		ConsensusData: block.ConsensusData{
			PrimaryIndex: 0,
			Nonce:        2083236893,
		},
	}

	if err = b.RebuildMerkleRoot(); err != nil {
		return nil, err
	}

	return b, nil
}

func deployNativeContracts() *transaction.Transaction {
	buf := io.NewBufBinWriter()
	emit.Syscall(buf.BinWriter, "Neo.Native.Deploy")
	script := buf.Bytes()
	tx := transaction.New(script, 0)
	tx.Nonce = 0
	tx.Sender = hash.Hash160([]byte{byte(opcode.PUSH1)})
	tx.Scripts = []transaction.Witness{
		{
			InvocationScript:   []byte{},
			VerificationScript: []byte{byte(opcode.PUSH1)},
		},
	}
	return tx
}

func getValidators(cfg config.ProtocolConfiguration) ([]*keys.PublicKey, error) {
	validators := make([]*keys.PublicKey, len(cfg.StandbyValidators))
	for i, pubKeyStr := range cfg.StandbyValidators {
		pubKey, err := keys.NewPublicKeyFromString(pubKeyStr)
		if err != nil {
			return nil, err
		}
		validators[i] = pubKey
	}
	return validators, nil
}

func getNextConsensusAddress(validators []*keys.PublicKey) (val util.Uint160, err error) {
	vlen := len(validators)
	raw, err := smartcontract.CreateMultiSigRedeemScript(
		vlen-(vlen-1)/3,
		validators,
	)
	if err != nil {
		return val, err
	}
	return hash.Hash160(raw), nil
}

func calculateUtilityAmount() util.Fixed8 {
	sum := 0
	for i := 0; i < len(genAmount); i++ {
		sum += genAmount[i]
	}
	return util.Fixed8FromInt64(int64(sum * decrementInterval))
}

// headerSliceReverse reverses the given slice of *Header.
func headerSliceReverse(dest []*block.Header) {
	for i, j := 0, len(dest)-1; i < j; i, j = i+1, j-1 {
		dest[i], dest[j] = dest[j], dest[i]
	}
}

// CalculateNetworkFee returns network fee for transaction
func CalculateNetworkFee(script []byte) (util.Fixed8, int) {
	var (
		netFee util.Fixed8
		size   int
	)
	if vm.IsSignatureContract(script) {
		size += 67 + io.GetVarSize(script)
		netFee = netFee.Add(opcodePrice(opcode.PUSHDATA1, opcode.PUSHNULL).Add(ecdsaVerifyInteropPrice))
	} else if n, pubs, ok := vm.ParseMultiSigContract(script); ok {
		m := len(pubs)
		sizeInv := 66 * m
		size += io.GetVarSize(sizeInv) + sizeInv + io.GetVarSize(script)
		netFee = netFee.Add(calculateMultisigFee(m)).Add(calculateMultisigFee(n))
		netFee = netFee.Add(opcodePrice(opcode.PUSHNULL)).Add(util.Fixed8(int64(ecdsaVerifyInteropPrice) * int64(n)))
	} else {
		// We can support more contract types in the future.
	}
	return netFee, size
}

func calculateMultisigFee(n int) util.Fixed8 {
	result := util.Fixed8(int64(opcodePrice(opcode.PUSHDATA1)) * int64(n))
	bw := io.NewBufBinWriter()
	emit.Int(bw.BinWriter, int64(n))
	// it's a hack because prices of small PUSH* opcodes are equal
	result = result.Add(opcodePrice(opcode.Opcode(bw.Bytes()[0])))
	return result
}
