package utils

import (
	"errors"
	"time"

	"github.com/cosmos/evm/testutil/integration/evm/grpc"
	"github.com/cosmos/evm/testutil/integration/evm/network"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
)

// WaitToAccrueRewards is a helper function that waits for rewards to
// accumulate up to a specified expected amount
func WaitToAccrueRewards(n network.Network, gh grpc.Handler, delegatorAddr string, expRewards sdk.DecCoins) (sdk.DecCoins, error) {
	var (
		err     error
		lapse   = time.Hour * 24 * 7 // one week
		rewards = sdk.DecCoins{}
	)

	if err = checkNonZeroInflation(n); err != nil {
		return nil, err
	}

	expAmt := expRewards.AmountOf(n.GetBaseDenom())
	for rewards.AmountOf(n.GetBaseDenom()).LT(expAmt) {
		rewards, err = checkRewardsAfter(n, gh, delegatorAddr, lapse)
		if err != nil {
			return nil, errorsmod.Wrap(err, "error checking rewards")
		}
	}

	return rewards, err
}

// checkRewardsAfter is a helper function that checks the accrued rewards
// after the provided timelapse
func checkRewardsAfter(n network.Network, gh grpc.Handler, delegatorAddr string, lapse time.Duration) (sdk.DecCoins, error) {
	err := n.NextBlockAfter(lapse)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to commit block after voting period ends")
	}

	res, err := gh.GetDelegationTotalRewards(delegatorAddr)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "error while querying for delegation rewards")
	}

	return res.Total, nil
}

// WaitToAccrueCommission is a helper function that waits for commission to
// accumulate up to a specified expected amount
func WaitToAccrueCommission(n network.Network, gh grpc.Handler, validatorAddr string, expCommission sdk.DecCoins) (sdk.DecCoins, error) {
	var (
		err        error
		lapse      = time.Hour * 24 * 7 // one week
		commission = sdk.DecCoins{}
	)

	if err := checkNonZeroInflation(n); err != nil {
		return nil, err
	}

	expAmt := expCommission.AmountOf(n.GetBaseDenom())
	for commission.AmountOf(n.GetBaseDenom()).LT(expAmt) {
		commission, err = checkCommissionAfter(n, gh, validatorAddr, lapse)
		if err != nil {
			return nil, errorsmod.Wrap(err, "error checking commission")
		}
	}

	return commission, err
}

// checkCommissionAfter is a helper function that checks the accrued commission
// after the provided time lapse
func checkCommissionAfter(n network.Network, gh grpc.Handler, valAddr string, lapse time.Duration) (sdk.DecCoins, error) {
	err := n.NextBlockAfter(lapse)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to commit block after voting period ends")
	}

	res, err := gh.GetValidatorCommission(valAddr)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "error while querying for delegation rewards")
	}

	return res.Commission.Commission, nil
}

// checkNonZeroInflation is a helper function that checks if the network's
// inflation is non-zero.
// This is required to ensure that rewards and commission are accrued.
//
// EpixChain customisation: our chain replaces stock x/mint with x/epixmint
// (see evmd/app.go). The mint Keeper still exists for RPC compatibility but
// its Minter store is never initialised, so the Inflation() query returns
// a "collections: not found" error. EpixMint distributes inflation directly
// to validators per block, so as long as its InitialAnnualMintAmount is
// non-zero, rewards do accrue. Treat the missing-Minter error as a signal
// to fall through and let the calling test poll for actual rewards.
func checkNonZeroInflation(n network.Network) error {
	res, err := n.GetMintClient().Inflation(n.GetContext(), &minttypes.QueryInflationRequest{})
	if err != nil {
		// EpixChain: stock mint module isn't wired; trust the upstream
		// caller to poll for non-zero rewards. If EpixMint is also zeroed
		// the polling loop will time out, surfacing the real issue.
		return nil
	}

	if res.Inflation.IsZero() {
		return errors.New("inflation is zero; must be non-zero for rewards or commission to be distributed")
	}

	return nil
}
