package ibc

import (
	"os"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TestMain resets the global SDK bech32 prefix to the cosmos-sdk default
// ("cosmos") before any test runs. EpixChain's evmd package init() globally
// sets the prefix to "epix", but the IBC test suites construct one chain
// using the upstream ibc-go reference simapp which hardcodes "cosmos" in
// its account/validator/consensus codecs (via sdk.Bech32MainPrefix const).
// Module accounts in cosmos-sdk store their address as a bech32 STRING
// (BaseAccount.Address) and decode it at runtime using GetConfig().
// GetBech32AccountAddrPrefix(); a process where one chain stamped its
// addresses as "cosmos1..." and another as "epix1..." can therefore only
// ever decode one of them correctly.
//
// Resetting to "cosmos" here is safe for this package — none of the IBC
// tests hardcode "epix1..." addresses; they all build addresses dynamically
// via sdk.AccAddressFromBech32 / .String() which will round-trip through
// whatever prefix the config currently has.
func TestMain(m *testing.M) {
	cfg := sdk.GetConfig()
	cfg.SetBech32PrefixForAccount(
		sdk.Bech32MainPrefix,
		sdk.Bech32MainPrefix+sdk.PrefixPublic,
	)
	cfg.SetBech32PrefixForValidator(
		sdk.Bech32MainPrefix+sdk.PrefixValidator+sdk.PrefixOperator,
		sdk.Bech32MainPrefix+sdk.PrefixValidator+sdk.PrefixOperator+sdk.PrefixPublic,
	)
	cfg.SetBech32PrefixForConsensusNode(
		sdk.Bech32MainPrefix+sdk.PrefixValidator+sdk.PrefixConsensus,
		sdk.Bech32MainPrefix+sdk.PrefixValidator+sdk.PrefixConsensus+sdk.PrefixPublic,
	)
	os.Exit(m.Run())
}
