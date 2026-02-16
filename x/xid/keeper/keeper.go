package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/evm/x/xid/types"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
)

// Keeper manages the xID module state
type Keeper struct {
	storeKey      storetypes.StoreKey
	cdc           codec.BinaryCodec
	authority     string
	bankKeeper    types.BankKeeper
	accountKeeper types.AccountKeeper
}

// NewKeeper creates a new xID keeper
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	authority string,
	bankKeeper types.BankKeeper,
	accountKeeper types.AccountKeeper,
) Keeper {
	return Keeper{
		storeKey:      storeKey,
		cdc:           cdc,
		authority:     authority,
		bankKeeper:    bankKeeper,
		accountKeeper: accountKeeper,
	}
}

// GetAuthority returns the module's authority address
func (k Keeper) GetAuthority() string {
	return k.authority
}

// Logger returns a module-specific logger
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}
