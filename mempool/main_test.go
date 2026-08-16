package mempool_test

import (
	"os"
	"testing"

	"github.com/cosmos/evm/testutil/constants"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TestMain configures the global SDK bech32 prefix to match the test
// constants before any test runs. EpixChain overrides ExampleBech32Prefix
// to "epix"; the cosmos-sdk default account address codec uses "cosmos",
// which causes sigTx.GetSigners() to fail with "hrp does not match bech32
// prefix" when tests construct addresses with the epix prefix.
//
// Configuring the global sdk Config here keeps the cosmos-sdk's address
// codec in sync with the prefix the test fixtures use.
func TestMain(m *testing.M) {
	cfg := sdk.GetConfig()
	cfg.SetBech32PrefixForAccount(
		constants.ExampleBech32Prefix,
		constants.ExampleBech32Prefix+sdk.PrefixPublic,
	)
	cfg.SetBech32PrefixForValidator(
		constants.ExampleBech32Prefix+sdk.PrefixValidator+sdk.PrefixOperator,
		constants.ExampleBech32Prefix+sdk.PrefixValidator+sdk.PrefixOperator+sdk.PrefixPublic,
	)
	cfg.SetBech32PrefixForConsensusNode(
		constants.ExampleBech32Prefix+sdk.PrefixValidator+sdk.PrefixConsensus,
		constants.ExampleBech32Prefix+sdk.PrefixValidator+sdk.PrefixConsensus+sdk.PrefixPublic,
	)
	os.Exit(m.Run())
}
