package core

import (
	"bytes"
	"errors"
	"fmt"
	"sort"

	"github.com/nspcc-dev/neo-go/pkg/core/interop"
	"github.com/nspcc-dev/neo-go/pkg/core/state"
	"github.com/nspcc-dev/neo-go/pkg/core/transaction"
	"github.com/nspcc-dev/neo-go/pkg/io"
	"github.com/nspcc-dev/neo-go/pkg/smartcontract/manifest"
	"github.com/nspcc-dev/neo-go/pkg/smartcontract/trigger"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/nspcc-dev/neo-go/pkg/vm"
	"github.com/nspcc-dev/neo-go/pkg/vm/stackitem"
)

const (
	// MaxContractDescriptionLen is the maximum length for contract description.
	MaxContractDescriptionLen = 65536
	// MaxContractScriptSize is the maximum script size for a contract.
	MaxContractScriptSize = 1024 * 1024
	// MaxContractParametersNum is the maximum number of parameters for a contract.
	MaxContractParametersNum = 252
	// MaxContractStringLen is the maximum length for contract metadata strings.
	MaxContractStringLen = 252
)

// headerGetVersion returns version from the header.
func headerGetVersion(ic *interop.Context, v *vm.VM) error {
	header, err := popHeaderFromVM(v)
	if err != nil {
		return err
	}
	v.Estack().PushVal(header.Version)
	return nil
}

// headerGetMerkleRoot returns version from the header.
func headerGetMerkleRoot(ic *interop.Context, v *vm.VM) error {
	header, err := popHeaderFromVM(v)
	if err != nil {
		return err
	}
	v.Estack().PushVal(header.MerkleRoot.BytesBE())
	return nil
}

// headerGetNextConsensus returns version from the header.
func headerGetNextConsensus(ic *interop.Context, v *vm.VM) error {
	header, err := popHeaderFromVM(v)
	if err != nil {
		return err
	}
	v.Estack().PushVal(header.NextConsensus.BytesBE())
	return nil
}

// txGetAttributes returns current transaction attributes.
func txGetAttributes(ic *interop.Context, v *vm.VM) error {
	txInterface := v.Estack().Pop().Value()
	tx, ok := txInterface.(*transaction.Transaction)
	if !ok {
		return errors.New("value is not a transaction")
	}
	if len(tx.Attributes) > vm.MaxArraySize {
		return errors.New("too many attributes")
	}
	attrs := make([]stackitem.Item, 0, len(tx.Attributes))
	for i := range tx.Attributes {
		attrs = append(attrs, stackitem.NewInterop(&tx.Attributes[i]))
	}
	v.Estack().PushVal(attrs)
	return nil
}

// txGetWitnesses returns current transaction witnesses.
func txGetWitnesses(ic *interop.Context, v *vm.VM) error {
	txInterface := v.Estack().Pop().Value()
	tx, ok := txInterface.(*transaction.Transaction)
	if !ok {
		return errors.New("value is not a transaction")
	}
	if len(tx.Scripts) > vm.MaxArraySize {
		return errors.New("too many outputs")
	}
	scripts := make([]stackitem.Item, 0, len(tx.Scripts))
	for i := range tx.Scripts {
		scripts = append(scripts, stackitem.NewInterop(&tx.Scripts[i]))
	}
	v.Estack().PushVal(scripts)
	return nil
}

// witnessGetVerificationScript returns current witness' script.
func witnessGetVerificationScript(ic *interop.Context, v *vm.VM) error {
	witInterface := v.Estack().Pop().Value()
	wit, ok := witInterface.(*transaction.Witness)
	if !ok {
		return errors.New("value is not a witness")
	}
	// It's important not to share wit.VerificationScript slice with the code running in VM.
	script := make([]byte, len(wit.VerificationScript))
	copy(script, wit.VerificationScript)
	v.Estack().PushVal(script)
	return nil
}

// attrGetData returns tx attribute data.
func attrGetData(ic *interop.Context, v *vm.VM) error {
	attrInterface := v.Estack().Pop().Value()
	attr, ok := attrInterface.(*transaction.Attribute)
	if !ok {
		return fmt.Errorf("%T is not an attribute", attr)
	}
	v.Estack().PushVal(attr.Data)
	return nil
}

// attrGetData returns tx attribute usage field.
func attrGetUsage(ic *interop.Context, v *vm.VM) error {
	attrInterface := v.Estack().Pop().Value()
	attr, ok := attrInterface.(*transaction.Attribute)
	if !ok {
		return fmt.Errorf("%T is not an attribute", attr)
	}
	v.Estack().PushVal(int(attr.Usage))
	return nil
}

// bcGetAccount returns or creates an account.
func bcGetAccount(ic *interop.Context, v *vm.VM) error {
	accbytes := v.Estack().Pop().Bytes()
	acchash, err := util.Uint160DecodeBytesBE(accbytes)
	if err != nil {
		return err
	}
	acc, err := ic.DAO.GetAccountStateOrNew(acchash)
	if err != nil {
		return err
	}
	v.Estack().PushVal(stackitem.NewInterop(acc))
	return nil
}

// accountGetBalance returns balance for a given account.
func accountGetBalance(ic *interop.Context, v *vm.VM) error {
	accInterface := v.Estack().Pop().Value()
	acc, ok := accInterface.(*state.Account)
	if !ok {
		return fmt.Errorf("%T is not an account state", acc)
	}
	asbytes := v.Estack().Pop().Bytes()
	ashash, err := util.Uint256DecodeBytesBE(asbytes)
	if err != nil {
		return err
	}
	balance, ok := acc.GetBalanceValues()[ashash]
	if !ok {
		balance = util.Fixed8(0)
	}
	v.Estack().PushVal(int64(balance))
	return nil
}

// accountGetScriptHash returns script hash of a given account.
func accountGetScriptHash(ic *interop.Context, v *vm.VM) error {
	accInterface := v.Estack().Pop().Value()
	acc, ok := accInterface.(*state.Account)
	if !ok {
		return fmt.Errorf("%T is not an account state", acc)
	}
	v.Estack().PushVal(acc.ScriptHash.BytesBE())
	return nil
}

// accountIsStandard checks whether given account is standard.
func accountIsStandard(ic *interop.Context, v *vm.VM) error {
	accbytes := v.Estack().Pop().Bytes()
	acchash, err := util.Uint160DecodeBytesBE(accbytes)
	if err != nil {
		return err
	}
	contract, err := ic.DAO.GetContractState(acchash)
	res := err != nil || vm.IsStandardContract(contract.Script)
	v.Estack().PushVal(res)
	return nil
}

// storageFind finds stored key-value pair.
func storageFind(ic *interop.Context, v *vm.VM) error {
	stcInterface := v.Estack().Pop().Value()
	stc, ok := stcInterface.(*StorageContext)
	if !ok {
		return fmt.Errorf("%T is not a StorageContext", stcInterface)
	}
	err := checkStorageContext(ic, stc)
	if err != nil {
		return err
	}
	prefix := v.Estack().Pop().Bytes()
	siMap, err := ic.DAO.GetStorageItemsWithPrefix(stc.ScriptHash, prefix)
	if err != nil {
		return err
	}

	filteredMap := stackitem.NewMap()
	for k, v := range siMap {
		filteredMap.Add(stackitem.NewByteArray(append(prefix, []byte(k)...)), stackitem.NewByteArray(v.Value))
	}
	sort.Slice(filteredMap.Value().([]stackitem.MapElement), func(i, j int) bool {
		return bytes.Compare(filteredMap.Value().([]stackitem.MapElement)[i].Key.Value().([]byte),
			filteredMap.Value().([]stackitem.MapElement)[j].Key.Value().([]byte)) == -1
	})

	item := vm.NewMapIterator(filteredMap)
	v.Estack().PushVal(item)

	return nil
}

// createContractStateFromVM pops all contract state elements from the VM
// evaluation stack, does a lot of checks and returns Contract if it
// succeeds.
func createContractStateFromVM(ic *interop.Context, v *vm.VM) (*state.Contract, error) {
	if ic.Trigger != trigger.Application {
		return nil, errors.New("can't create contract when not triggered by an application")
	}
	script := v.Estack().Pop().Bytes()
	if len(script) > MaxContractScriptSize {
		return nil, errors.New("the script is too big")
	}
	manifestBytes := v.Estack().Pop().Bytes()
	if len(manifestBytes) > manifest.MaxManifestSize {
		return nil, errors.New("manifest is too big")
	}
	if !v.AddGas(util.Fixed8(StoragePrice * (len(script) + len(manifestBytes)))) {
		return nil, errors.New("gas limit exceeded")
	}
	var m manifest.Manifest
	r := io.NewBinReaderFromBuf(manifestBytes)
	m.DecodeBinary(r)
	if r.Err != nil {
		return nil, r.Err
	}
	return &state.Contract{
		Script:   script,
		Manifest: m,
	}, nil
}

// contractCreate creates a contract.
func contractCreate(ic *interop.Context, v *vm.VM) error {
	newcontract, err := createContractStateFromVM(ic, v)
	if err != nil {
		return err
	}
	contract, err := ic.DAO.GetContractState(newcontract.ScriptHash())
	if contract != nil {
		return errors.New("contract already exists")
	} else if err := ic.DAO.PutContractState(newcontract); err != nil {
		return err
	}
	v.Estack().PushVal(stackitem.NewInterop(newcontract))
	return nil
}

// contractGetScript returns a script associated with a contract.
func contractGetScript(ic *interop.Context, v *vm.VM) error {
	csInterface := v.Estack().Pop().Value()
	cs, ok := csInterface.(*state.Contract)
	if !ok {
		return fmt.Errorf("%T is not a contract state", cs)
	}
	v.Estack().PushVal(cs.Script)
	return nil
}

// contractIsPayable returns whether contract is payable.
func contractIsPayable(ic *interop.Context, v *vm.VM) error {
	csInterface := v.Estack().Pop().Value()
	cs, ok := csInterface.(*state.Contract)
	if !ok {
		return fmt.Errorf("%T is not a contract state", cs)
	}
	v.Estack().PushVal(cs.IsPayable())
	return nil
}

// contractMigrate migrates a contract.
func contractMigrate(ic *interop.Context, v *vm.VM) error {
	contract, err := ic.DAO.GetContractState(v.GetCurrentScriptHash())
	if contract == nil {
		return errors.New("contract doesn't exist")
	}
	newcontract, err := createContractStateFromVM(ic, v)
	if err != nil {
		return err
	}
	if newcontract.Script != nil {
		if l := len(newcontract.Script); l == 0 || l > MaxContractScriptSize {
			return errors.New("invalid script len")
		}
		h := newcontract.ScriptHash()
		if h.Equals(contract.ScriptHash()) {
			return errors.New("the script is the same")
		} else if _, err := ic.DAO.GetContractState(h); err == nil {
			return errors.New("contract already exists")
		}
		newcontract.ID = contract.ID
		if err := ic.DAO.PutContractState(newcontract); err != nil {
			return err
		}
		if err := ic.DAO.DeleteContractState(contract.ScriptHash()); err != nil {
			return err
		}
	}
	if contract.HasStorage() {
		// TODO store items by ID #1037
		hash := v.GetCurrentScriptHash()
		siMap, err := ic.DAO.GetStorageItems(hash)
		if err != nil {
			return err
		}
		for k, v := range siMap {
			v.IsConst = false
			err = ic.DAO.PutStorageItem(contract.ScriptHash(), []byte(k), v)
			if err != nil {
				return err
			}
		}
	}
	v.Estack().PushVal(stackitem.NewInterop(contract))
	return contractDestroy(ic, v)
}

// runtimeSerialize serializes top stack item into a ByteArray.
func runtimeSerialize(_ *interop.Context, v *vm.VM) error {
	return vm.RuntimeSerialize(v)
}

// runtimeDeserialize deserializes ByteArray from a stack into an item.
func runtimeDeserialize(_ *interop.Context, v *vm.VM) error {
	return vm.RuntimeDeserialize(v)
}
