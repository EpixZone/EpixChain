package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"

	"github.com/cosmos/evm/x/vrf/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// EndBlock computes and stores a new random beacon each block.
// beacon[N] = SHA256(beacon[N-1] || blockHash[N] || proposerAddr[N] || blockTime[N] || dataHash[N])
//
// dataHash is the Merkle root of all transactions in the block. Including it
// means the beacon depends on every transaction hash, making it practically
// impossible for a block proposer to predict or manipulate the outcome —
// they would need to control all user-submitted transactions in the block.
func (k Keeper) EndBlock(ctx sdk.Context) error {
	params := k.GetParams(ctx)
	if !params.Enabled {
		return nil
	}

	height := uint64(ctx.BlockHeight())

	// Get previous beacon (empty string for genesis block)
	prevBeacon := ""
	if prev, found := k.GetLatestBeacon(ctx); found {
		prevBeacon = prev.Beacon
	}

	// Compute new beacon from block data
	header := ctx.BlockHeader()
	proposer := sdk.ConsAddress(header.ProposerAddress).String()
	blockHash := hex.EncodeToString(ctx.HeaderHash())
	blockTime := strconv.FormatInt(ctx.BlockTime().Unix(), 10)
	dataHash := hex.EncodeToString(header.DataHash)

	h := sha256.New()
	h.Write([]byte(prevBeacon))
	h.Write([]byte(blockHash))
	h.Write([]byte(proposer))
	h.Write([]byte(blockTime))
	h.Write([]byte(dataHash))
	beacon := hex.EncodeToString(h.Sum(nil))

	record := types.RandomBeacon{
		Height:    height,
		Beacon:    beacon,
		Proposer:  proposer,
		Timestamp: ctx.BlockTime().Unix(),
	}
	k.SetBeacon(ctx, record)
	k.SetLatestHeight(ctx, height)

	// Prune old beacons beyond the lookback window
	if height > params.LookbackBlocks {
		k.PruneBeaconsBelow(ctx, height-params.LookbackBlocks)
	}

	return nil
}
