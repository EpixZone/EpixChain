package vrf

import (
	"bytes"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"

	_ "embed"

	cmn "github.com/cosmos/evm/precompiles/common"
	evmtypes "github.com/cosmos/evm/x/vm/types"
	"github.com/cosmos/evm/x/vrf/keeper"

	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Method names matching the ABI
const (
	GetBeaconMethod           = "getBeacon"
	LatestBeaconMethod        = "latestBeacon"
	GetMultiBlockBeaconMethod = "getMultiBlockBeacon"
)

var _ vm.PrecompiledContract = &Precompile{}

var (
	//go:embed abi.json
	f   []byte
	ABI abi.ABI
)

func init() {
	var err error
	ABI, err = abi.JSON(bytes.NewReader(f))
	if err != nil {
		panic(err)
	}
}

// Precompile defines the VRF precompile for EVM access
type Precompile struct {
	cmn.Precompile

	abi.ABI
	vrfKeeper keeper.Keeper
}

// NewPrecompile creates a new VRF Precompile instance
func NewPrecompile(
	vrfKeeper keeper.Keeper,
) *Precompile {
	return &Precompile{
		Precompile: cmn.Precompile{
			KvGasConfig:          storetypes.KVGasConfig(),
			TransientKVGasConfig: storetypes.TransientGasConfig(),
			ContractAddress:      common.HexToAddress(evmtypes.VRFPrecompileAddress),
		},
		ABI:       ABI,
		vrfKeeper: vrfKeeper,
	}
}

// RequiredGas returns the required bare minimum gas to execute the precompile.
func (p Precompile) RequiredGas(input []byte) uint64 {
	if len(input) < 4 {
		return 0
	}

	methodID := input[:4]
	method, err := p.MethodById(methodID)
	if err != nil {
		return 0
	}

	return p.Precompile.RequiredGas(input, p.IsTransaction(method))
}

// Name returns the name of the precompile, used by go-ethereum's PrecompiledContract interface.
func (Precompile) Name() string {
	return "vrf"
}

// Run executes the precompile
func (p Precompile) Run(evm *vm.EVM, contract *vm.Contract, readonly bool) ([]byte, error) {
	return p.RunNativeAction(evm, contract, func(ctx sdk.Context) ([]byte, error) {
		return p.Execute(ctx, contract, readonly)
	})
}

// Execute dispatches to the appropriate method handler
func (p Precompile) Execute(ctx sdk.Context, contract *vm.Contract, readOnly bool) ([]byte, error) {
	method, args, err := cmn.SetupABI(p.ABI, contract, readOnly, p.IsTransaction)
	if err != nil {
		return nil, err
	}

	var bz []byte

	switch method.Name {
	case GetBeaconMethod:
		bz, err = p.GetBeacon(ctx, method, args)
	case LatestBeaconMethod:
		bz, err = p.LatestBeacon(ctx, method, args)
	case GetMultiBlockBeaconMethod:
		bz, err = p.GetMultiBlockBeacon(ctx, method, args)
	default:
		return nil, fmt.Errorf(cmn.ErrUnknownMethod, method.Name)
	}

	return bz, err
}

// IsTransaction checks if the given method is a state-changing transaction.
// VRF precompile is read-only, so this always returns false.
func (Precompile) IsTransaction(method *abi.Method) bool {
	return false
}
