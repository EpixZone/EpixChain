package vrf

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetBeacon handles the getBeacon(blockHeight) view function.
// Returns the beacon as bytes32.
func (p Precompile) GetBeacon(
	ctx sdk.Context,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("expected 1 argument, got %d", len(args))
	}

	height, ok := args[0].(uint64)
	if !ok {
		// Try *big.Int fallback
		if bi, ok2 := args[0].(*big.Int); ok2 {
			height = bi.Uint64()
		} else {
			return nil, fmt.Errorf("invalid argument type for blockHeight: %T", args[0])
		}
	}

	beacon, found := p.vrfKeeper.GetBeaconByHeight(ctx, height)
	if !found {
		// Return zero bytes32 if not found
		return method.Outputs.Pack([32]byte{})
	}

	beaconBytes, err := hexToBytes32(beacon.Beacon)
	if err != nil {
		return nil, err
	}

	return method.Outputs.Pack(beaconBytes)
}

// LatestBeacon handles the latestBeacon() view function.
// Returns the beacon as bytes32 and the block height.
func (p Precompile) LatestBeacon(
	ctx sdk.Context,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	beacon, found := p.vrfKeeper.GetLatestBeacon(ctx)
	if !found {
		return method.Outputs.Pack([32]byte{}, uint64(0))
	}

	beaconBytes, err := hexToBytes32(beacon.Beacon)
	if err != nil {
		return nil, err
	}

	return method.Outputs.Pack(beaconBytes, beacon.Height)
}

// GetMultiBlockBeacon handles getMultiBlockBeacon(endHeight, blocks).
// It hashes together N consecutive beacons ending at endHeight, so that
// every block proposer in the range contributes entropy. To manipulate
// the result, ALL proposers across all N blocks would need to collude.
func (p Precompile) GetMultiBlockBeacon(
	ctx sdk.Context,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("expected 2 arguments, got %d", len(args))
	}

	endHeight, ok := args[0].(uint64)
	if !ok {
		if bi, ok2 := args[0].(*big.Int); ok2 {
			endHeight = bi.Uint64()
		} else {
			return nil, fmt.Errorf("invalid argument type for endHeight: %T", args[0])
		}
	}

	blocks, ok := args[1].(uint64)
	if !ok {
		if bi, ok2 := args[1].(*big.Int); ok2 {
			blocks = bi.Uint64()
		} else {
			return nil, fmt.Errorf("invalid argument type for blocks: %T", args[1])
		}
	}

	if blocks == 0 || blocks > 256 {
		return nil, fmt.Errorf("blocks must be between 1 and 256")
	}
	if endHeight < blocks {
		return nil, fmt.Errorf("endHeight must be >= blocks")
	}

	startHeight := endHeight - blocks + 1

	// Hash all beacons in the range together
	h := sha256.New()
	for height := startHeight; height <= endHeight; height++ {
		beacon, found := p.vrfKeeper.GetBeaconByHeight(ctx, height)
		if !found {
			return method.Outputs.Pack([32]byte{})
		}
		h.Write([]byte(beacon.Beacon))
	}

	var result [32]byte
	copy(result[:], h.Sum(nil))
	return method.Outputs.Pack(result)
}

// hexToBytes32 converts a 64-char hex string to a [32]byte array
func hexToBytes32(hexStr string) ([32]byte, error) {
	var result [32]byte
	if len(hexStr) != 64 {
		return result, fmt.Errorf("expected 64-char hex string, got %d chars", len(hexStr))
	}
	decoded, err := hex.DecodeString(hexStr)
	if err != nil {
		return result, err
	}
	copy(result[:], decoded)
	return result, nil
}
