package keeper_test

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strconv"
	"time"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"github.com/cosmos/evm/x/vrf/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ctxWithBlock builds a context with realistic block header data for EndBlock testing.
func (s *KeeperTestSuite) ctxWithBlock(height int64, proposer []byte, blockHash []byte, dataHash []byte, blockTime time.Time) sdk.Context {
	return s.ctx.
		WithBlockHeight(height).
		WithBlockHeader(cmtproto.Header{
			Height:          height,
			ProposerAddress: proposer,
			DataHash:        dataHash,
			Time:            blockTime,
		}).
		WithHeaderHash(blockHash)
}

// computeExpectedBeacon replicates the beacon computation from abci.go
func computeExpectedBeacon(prevBeacon string, blockHash []byte, proposer []byte, blockTime time.Time, dataHash []byte) string {
	proposerStr := sdk.ConsAddress(proposer).String()
	blockHashHex := hex.EncodeToString(blockHash)
	blockTimeStr := strconv.FormatInt(blockTime.Unix(), 10)
	dataHashHex := hex.EncodeToString(dataHash)

	h := sha256.New()
	h.Write([]byte(prevBeacon))
	h.Write([]byte(blockHashHex))
	h.Write([]byte(proposerStr))
	h.Write([]byte(blockTimeStr))
	h.Write([]byte(dataHashHex))
	return hex.EncodeToString(h.Sum(nil))
}

func (s *KeeperTestSuite) TestEndBlock_FirstBlock() {
	err := s.keeper.SetParams(s.ctx, s.keeper.GetParams(s.ctx)) // ensure defaults
	s.Require().NoError(err)

	proposer := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
	blockHash := []byte{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0x00, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89}
	dataHash := []byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00}
	blockTime := time.Unix(1700000000, 0)

	ctx := s.ctxWithBlock(1, proposer, blockHash, dataHash, blockTime)
	err = s.keeper.EndBlock(ctx)
	s.Require().NoError(err)

	// Verify beacon was stored
	beacon, found := s.keeper.GetBeaconByHeight(ctx, 1)
	s.Require().True(found)
	s.Require().Equal(uint64(1), beacon.Height)
	s.Require().Equal(int64(1700000000), beacon.Timestamp)

	// Verify latest height
	h, found := s.keeper.GetLatestHeight(ctx)
	s.Require().True(found)
	s.Require().Equal(uint64(1), h)

	// Manually compute and compare beacon value
	expected := computeExpectedBeacon("", blockHash, proposer, blockTime, dataHash)
	s.Require().Equal(expected, beacon.Beacon)
}

func (s *KeeperTestSuite) TestEndBlock_ChainedBeacons() {
	proposer := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
	blockHash1 := make([]byte, 32)
	blockHash1[0] = 0xAA
	dataHash1 := make([]byte, 32)
	dataHash1[0] = 0x11
	t1 := time.Unix(1000, 0)

	blockHash2 := make([]byte, 32)
	blockHash2[0] = 0xBB
	dataHash2 := make([]byte, 32)
	dataHash2[0] = 0x22
	t2 := time.Unix(2000, 0)

	// Block 1
	ctx1 := s.ctxWithBlock(1, proposer, blockHash1, dataHash1, t1)
	err := s.keeper.EndBlock(ctx1)
	s.Require().NoError(err)

	beacon1, found := s.keeper.GetBeaconByHeight(ctx1, 1)
	s.Require().True(found)

	// Block 2 — should chain from beacon 1
	ctx2 := s.ctxWithBlock(2, proposer, blockHash2, dataHash2, t2)
	err = s.keeper.EndBlock(ctx2)
	s.Require().NoError(err)

	beacon2, found := s.keeper.GetBeaconByHeight(ctx2, 2)
	s.Require().True(found)

	// Verify chaining: beacon2 uses beacon1 as prevBeacon
	expected := computeExpectedBeacon(beacon1.Beacon, blockHash2, proposer, t2, dataHash2)
	s.Require().Equal(expected, beacon2.Beacon)

	// Beacons should be different
	s.Require().NotEqual(beacon1.Beacon, beacon2.Beacon)
}

func (s *KeeperTestSuite) TestEndBlock_Deterministic() {
	proposer := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
	blockHash := make([]byte, 32)
	blockHash[0] = 0xAA
	dataHash := make([]byte, 32)
	dataHash[0] = 0x11
	bt := time.Unix(1000, 0)

	// First run
	ctx1 := s.ctxWithBlock(1, proposer, blockHash, dataHash, bt)
	err := s.keeper.EndBlock(ctx1)
	s.Require().NoError(err)
	beacon1, _ := s.keeper.GetBeaconByHeight(ctx1, 1)

	// Reset and second run with identical inputs
	s.SetupTest()
	ctx2 := s.ctxWithBlock(1, proposer, blockHash, dataHash, bt)
	err = s.keeper.EndBlock(ctx2)
	s.Require().NoError(err)
	beacon2, _ := s.keeper.GetBeaconByHeight(ctx2, 1)

	s.Require().Equal(beacon1.Beacon, beacon2.Beacon)
}

func (s *KeeperTestSuite) TestEndBlock_DifferentInputsProduceDifferentBeacons() {
	proposer := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
	blockHash := make([]byte, 32)
	blockHash[0] = 0xAA
	dataHash := make([]byte, 32)
	dataHash[0] = 0x11
	bt := time.Unix(1000, 0)

	// Base beacon
	ctx1 := s.ctxWithBlock(1, proposer, blockHash, dataHash, bt)
	err := s.keeper.EndBlock(ctx1)
	s.Require().NoError(err)
	base, _ := s.keeper.GetBeaconByHeight(ctx1, 1)

	// Different blockHash
	s.SetupTest()
	blockHash2 := make([]byte, 32)
	blockHash2[0] = 0xBB
	ctx2 := s.ctxWithBlock(1, proposer, blockHash2, dataHash, bt)
	err = s.keeper.EndBlock(ctx2)
	s.Require().NoError(err)
	diffHash, _ := s.keeper.GetBeaconByHeight(ctx2, 1)
	s.Require().NotEqual(base.Beacon, diffHash.Beacon, "different blockHash should produce different beacon")

	// Different proposer
	s.SetupTest()
	proposer2 := []byte{20, 19, 18, 17, 16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}
	ctx3 := s.ctxWithBlock(1, proposer2, blockHash, dataHash, bt)
	err = s.keeper.EndBlock(ctx3)
	s.Require().NoError(err)
	diffProp, _ := s.keeper.GetBeaconByHeight(ctx3, 1)
	s.Require().NotEqual(base.Beacon, diffProp.Beacon, "different proposer should produce different beacon")

	// Different blockTime
	s.SetupTest()
	ctx4 := s.ctxWithBlock(1, proposer, blockHash, dataHash, time.Unix(9999, 0))
	err = s.keeper.EndBlock(ctx4)
	s.Require().NoError(err)
	diffTime, _ := s.keeper.GetBeaconByHeight(ctx4, 1)
	s.Require().NotEqual(base.Beacon, diffTime.Beacon, "different blockTime should produce different beacon")

	// Different dataHash
	s.SetupTest()
	dataHash2 := make([]byte, 32)
	dataHash2[0] = 0x99
	ctx5 := s.ctxWithBlock(1, proposer, blockHash, dataHash2, bt)
	err = s.keeper.EndBlock(ctx5)
	s.Require().NoError(err)
	diffData, _ := s.keeper.GetBeaconByHeight(ctx5, 1)
	s.Require().NotEqual(base.Beacon, diffData.Beacon, "different dataHash should produce different beacon")
}

func (s *KeeperTestSuite) TestEndBlock_DisabledNoOp() {
	err := s.keeper.SetParams(s.ctx, types.Params{Enabled: false, LookbackBlocks: 256})
	s.Require().NoError(err)

	proposer := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
	ctx := s.ctxWithBlock(1, proposer, make([]byte, 32), make([]byte, 32), time.Unix(1000, 0))
	err = s.keeper.EndBlock(ctx)
	s.Require().NoError(err)

	_, found := s.keeper.GetLatestBeacon(ctx)
	s.Require().False(found, "no beacon should be stored when disabled")
}

func (s *KeeperTestSuite) TestEndBlock_PruningTriggered() {
	err := s.keeper.SetParams(s.ctx, types.Params{Enabled: true, LookbackBlocks: 3})
	s.Require().NoError(err)

	proposer := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}

	// Run EndBlock for heights 1-6
	for h := int64(1); h <= 6; h++ {
		blockHash := make([]byte, 32)
		blockHash[0] = byte(h)
		dataHash := make([]byte, 32)
		dataHash[0] = byte(h + 100)
		ctx := s.ctxWithBlock(h, proposer, blockHash, dataHash, time.Unix(h*1000, 0))
		err := s.keeper.EndBlock(ctx)
		s.Require().NoError(err)
	}

	// After height 6 with lookback=3: prune below 6-3=3
	// Heights 1, 2 should be gone (iterator is [prefix, BeaconKey(3)) exclusive)
	for _, h := range []uint64{1, 2} {
		_, found := s.keeper.GetBeaconByHeight(s.ctx, h)
		s.Require().False(found, "height %d should be pruned", h)
	}

	// Heights 3-6 should remain
	for _, h := range []uint64{3, 4, 5, 6} {
		_, found := s.keeper.GetBeaconByHeight(s.ctx, h)
		s.Require().True(found, "height %d should remain", h)
	}
}

func (s *KeeperTestSuite) TestEndBlock_BeaconFormat() {
	proposer := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
	ctx := s.ctxWithBlock(1, proposer, make([]byte, 32), make([]byte, 32), time.Unix(1000, 0))
	err := s.keeper.EndBlock(ctx)
	s.Require().NoError(err)

	beacon, found := s.keeper.GetBeaconByHeight(ctx, 1)
	s.Require().True(found)

	// 64 lowercase hex chars
	s.Require().Len(beacon.Beacon, 64)
	s.Require().Regexp(regexp.MustCompile(`^[0-9a-f]{64}$`), beacon.Beacon)
}

func (s *KeeperTestSuite) TestEndBlock_TimestampCorrect() {
	proposer := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
	bt := time.Unix(1700000000, 0)
	ctx := s.ctxWithBlock(1, proposer, make([]byte, 32), make([]byte, 32), bt)
	err := s.keeper.EndBlock(ctx)
	s.Require().NoError(err)

	beacon, _ := s.keeper.GetBeaconByHeight(ctx, 1)
	s.Require().Equal(int64(1700000000), beacon.Timestamp)
}
