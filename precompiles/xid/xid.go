package xid

import (
	"bytes"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"

	_ "embed"

	cmn "github.com/cosmos/evm/precompiles/common"
	evmtypes "github.com/cosmos/evm/x/vm/types"
	"github.com/cosmos/evm/x/xid/keeper"

	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Method names matching the ABI
const (
	// Transactions
	RegisterMethod        = "register"
	TransferNameMethod    = "transferName"
	UpdateProfileMethod   = "updateProfile"
	SetDNSRecordMethod    = "setDNSRecord"
	DeleteDNSRecordMethod = "deleteDNSRecord"

	SetEpixNetPeerMethod    = "setEpixNetPeer"
	DeleteEpixNetPeerMethod = "deleteEpixNetPeer"

	// Queries
	ResolveMethod            = "resolve"
	ReverseResolveMethod     = "reverseResolve"
	GetProfileMethod         = "getProfile"
	GetDNSRecordMethod       = "getDNSRecord"
	GetRegistrationFeeMethod = "getRegistrationFee"
	GetEpixNetPeersMethod    = "getEpixNetPeers"
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

// Precompile defines the xID precompile for EVM access
type Precompile struct {
	cmn.Precompile

	abi.ABI
	xidKeeper  keeper.Keeper
	bankKeeper cmn.BankKeeper
}

// NewPrecompile creates a new xID Precompile instance
func NewPrecompile(
	xidKeeper keeper.Keeper,
	bankKeeper cmn.BankKeeper,
) *Precompile {
	return &Precompile{
		Precompile: cmn.Precompile{
			KvGasConfig:           storetypes.KVGasConfig(),
			TransientKVGasConfig:  storetypes.TransientGasConfig(),
			ContractAddress:       common.HexToAddress(evmtypes.XIDPrecompileAddress),
			BalanceHandlerFactory: cmn.NewBalanceHandlerFactory(bankKeeper),
		},
		ABI:        ABI,
		xidKeeper:  xidKeeper,
		bankKeeper: bankKeeper,
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

// Run executes the precompile
func (p Precompile) Run(evm *vm.EVM, contract *vm.Contract, readonly bool) ([]byte, error) {
	return p.RunNativeAction(evm, contract, func(ctx sdk.Context) ([]byte, error) {
		return p.Execute(ctx, evm.StateDB, contract, readonly)
	})
}

// Execute dispatches to the appropriate method handler
func (p Precompile) Execute(ctx sdk.Context, stateDB vm.StateDB, contract *vm.Contract, readOnly bool) ([]byte, error) {
	method, args, err := cmn.SetupABI(p.ABI, contract, readOnly, p.IsTransaction)
	if err != nil {
		return nil, err
	}

	var bz []byte

	switch method.Name {
	// Transactions
	case RegisterMethod:
		bz, err = p.Register(ctx, contract, stateDB, method, args)
	case TransferNameMethod:
		bz, err = p.TransferName(ctx, contract, stateDB, method, args)
	case UpdateProfileMethod:
		bz, err = p.UpdateProfile(ctx, contract, stateDB, method, args)
	case SetDNSRecordMethod:
		bz, err = p.SetDNSRecord(ctx, contract, stateDB, method, args)
	case DeleteDNSRecordMethod:
		bz, err = p.DeleteDNSRecord(ctx, contract, stateDB, method, args)
	case SetEpixNetPeerMethod:
		bz, err = p.SetEpixNetPeer(ctx, contract, stateDB, method, args)
	case DeleteEpixNetPeerMethod:
		bz, err = p.DeleteEpixNetPeer(ctx, contract, stateDB, method, args)

	// Queries
	case ResolveMethod:
		bz, err = p.Resolve(ctx, method, args)
	case ReverseResolveMethod:
		bz, err = p.ReverseResolve(ctx, method, args)
	case GetProfileMethod:
		bz, err = p.GetProfile(ctx, method, args)
	case GetDNSRecordMethod:
		bz, err = p.GetDNSRecord(ctx, method, args)
	case GetRegistrationFeeMethod:
		bz, err = p.GetRegistrationFee(ctx, method, args)
	case GetEpixNetPeersMethod:
		bz, err = p.GetEpixNetPeers(ctx, method, args)
	default:
		return nil, fmt.Errorf(cmn.ErrUnknownMethod, method.Name)
	}

	return bz, err
}

// IsTransaction checks if the given method is a state-changing transaction
func (Precompile) IsTransaction(method *abi.Method) bool {
	switch method.Name {
	case RegisterMethod, TransferNameMethod, UpdateProfileMethod, SetDNSRecordMethod, DeleteDNSRecordMethod, SetEpixNetPeerMethod, DeleteEpixNetPeerMethod:
		return true
	default:
		return false
	}
}
