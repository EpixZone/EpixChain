package keeper_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/evm/x/vrf/keeper"
	"github.com/cosmos/evm/x/vrf/types"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

type KeeperTestSuite struct {
	suite.Suite

	ctx    sdk.Context
	keeper keeper.Keeper
	cdc    codec.BinaryCodec
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) SetupTest() {
	encCfg := moduletestutil.MakeTestEncodingConfig()
	key := storetypes.NewKVStoreKey(types.StoreKey)
	testCtx := testutil.DefaultContextWithDB(s.T(), key, storetypes.NewTransientStoreKey("transient_test"))
	s.ctx = testCtx.Ctx
	s.cdc = encCfg.Codec
	s.keeper = keeper.NewKeeper(
		s.cdc,
		key,
		authtypes.NewModuleAddress("gov").String(),
	)
}

func makeBeacon(height uint64) types.RandomBeacon {
	return types.RandomBeacon{
		Height:    height,
		Beacon:    strings.Repeat("ab", 32),
		Proposer:  "cosmosvalcons1test",
		Timestamp: int64(height) * 1000, //nolint:gosec // G115
	}
}

// ---- State CRUD tests ----

func (s *KeeperTestSuite) TestSetAndGetBeacon() {
	beacon := makeBeacon(100)
	s.keeper.SetBeacon(s.ctx, beacon)

	got, found := s.keeper.GetBeaconByHeight(s.ctx, 100)
	s.Require().True(found)
	s.Require().Equal(beacon.Height, got.Height)
	s.Require().Equal(beacon.Beacon, got.Beacon)
	s.Require().Equal(beacon.Proposer, got.Proposer)
	s.Require().Equal(beacon.Timestamp, got.Timestamp)

	// Missing height
	_, found = s.keeper.GetBeaconByHeight(s.ctx, 999)
	s.Require().False(found)
}

func (s *KeeperTestSuite) TestSetBeacon_Overwrite() {
	b1 := types.RandomBeacon{Height: 100, Beacon: strings.Repeat("aa", 32), Proposer: "val1", Timestamp: 1000}
	b2 := types.RandomBeacon{Height: 100, Beacon: strings.Repeat("bb", 32), Proposer: "val2", Timestamp: 2000}

	s.keeper.SetBeacon(s.ctx, b1)
	s.keeper.SetBeacon(s.ctx, b2)

	got, found := s.keeper.GetBeaconByHeight(s.ctx, 100)
	s.Require().True(found)
	s.Require().Equal(b2.Beacon, got.Beacon)
	s.Require().Equal(b2.Proposer, got.Proposer)
}

func (s *KeeperTestSuite) TestGetLatestBeacon_Empty() {
	_, found := s.keeper.GetLatestBeacon(s.ctx)
	s.Require().False(found)
}

func (s *KeeperTestSuite) TestGetLatestBeacon_WithData() {
	s.keeper.SetBeacon(s.ctx, makeBeacon(10))
	s.keeper.SetBeacon(s.ctx, makeBeacon(20))
	s.keeper.SetBeacon(s.ctx, makeBeacon(30))
	s.keeper.SetLatestHeight(s.ctx, 30)

	got, found := s.keeper.GetLatestBeacon(s.ctx)
	s.Require().True(found)
	s.Require().Equal(uint64(30), got.Height)
}

func (s *KeeperTestSuite) TestSetAndGetLatestHeight() {
	// Fresh store
	height, found := s.keeper.GetLatestHeight(s.ctx)
	s.Require().False(found)
	s.Require().Equal(uint64(0), height)

	// Set and get
	s.keeper.SetLatestHeight(s.ctx, 42)
	height, found = s.keeper.GetLatestHeight(s.ctx)
	s.Require().True(found)
	s.Require().Equal(uint64(42), height)

	// Overwrite
	s.keeper.SetLatestHeight(s.ctx, 100)
	height, found = s.keeper.GetLatestHeight(s.ctx)
	s.Require().True(found)
	s.Require().Equal(uint64(100), height)
}

func (s *KeeperTestSuite) TestPruneBeaconsBelow() {
	for _, h := range []uint64{1, 2, 3, 4, 5, 10, 20} {
		s.keeper.SetBeacon(s.ctx, makeBeacon(h))
	}

	s.keeper.PruneBeaconsBelow(s.ctx, 5)

	// Heights 1-4 should be gone
	for _, h := range []uint64{1, 2, 3, 4} {
		_, found := s.keeper.GetBeaconByHeight(s.ctx, h)
		s.Require().False(found, "height %d should be pruned", h)
	}

	// Heights 5, 10, 20 should remain
	for _, h := range []uint64{5, 10, 20} {
		_, found := s.keeper.GetBeaconByHeight(s.ctx, h)
		s.Require().True(found, "height %d should remain", h)
	}
}

func (s *KeeperTestSuite) TestPruneBeaconsBelow_NothingToPrune() {
	s.keeper.SetBeacon(s.ctx, makeBeacon(100))
	s.keeper.SetBeacon(s.ctx, makeBeacon(200))

	s.keeper.PruneBeaconsBelow(s.ctx, 50)

	_, found := s.keeper.GetBeaconByHeight(s.ctx, 100)
	s.Require().True(found)
	_, found = s.keeper.GetBeaconByHeight(s.ctx, 200)
	s.Require().True(found)
}

func (s *KeeperTestSuite) TestPruneBeaconsBelow_PruneAll() {
	s.keeper.SetBeacon(s.ctx, makeBeacon(1))
	s.keeper.SetBeacon(s.ctx, makeBeacon(2))
	s.keeper.SetBeacon(s.ctx, makeBeacon(3))

	s.keeper.PruneBeaconsBelow(s.ctx, 100)

	for _, h := range []uint64{1, 2, 3} {
		_, found := s.keeper.GetBeaconByHeight(s.ctx, h)
		s.Require().False(found, "height %d should be pruned", h)
	}
}

func (s *KeeperTestSuite) TestIterateBeacons() {
	s.keeper.SetBeacon(s.ctx, makeBeacon(15))
	s.keeper.SetBeacon(s.ctx, makeBeacon(5))
	s.keeper.SetBeacon(s.ctx, makeBeacon(10))

	var collected []types.RandomBeacon
	s.keeper.IterateBeacons(s.ctx, func(beacon types.RandomBeacon) bool {
		collected = append(collected, beacon)
		return false
	})

	s.Require().Len(collected, 3)
	// Should be ascending order due to big-endian key encoding
	s.Require().Equal(uint64(5), collected[0].Height)
	s.Require().Equal(uint64(10), collected[1].Height)
	s.Require().Equal(uint64(15), collected[2].Height)
}

func (s *KeeperTestSuite) TestIterateBeacons_EarlyStop() {
	for _, h := range []uint64{1, 2, 3, 4, 5} {
		s.keeper.SetBeacon(s.ctx, makeBeacon(h))
	}

	var collected []types.RandomBeacon
	s.keeper.IterateBeacons(s.ctx, func(beacon types.RandomBeacon) bool {
		collected = append(collected, beacon)
		return len(collected) >= 2 // stop after 2
	})

	s.Require().Len(collected, 2)
}

func (s *KeeperTestSuite) TestIterateBeacons_Empty() {
	var collected []types.RandomBeacon
	s.keeper.IterateBeacons(s.ctx, func(beacon types.RandomBeacon) bool {
		collected = append(collected, beacon)
		return false
	})

	s.Require().Empty(collected)
}

// ---- Params tests ----

func (s *KeeperTestSuite) TestSetAndGetParams() {
	// Default before setting
	params := s.keeper.GetParams(s.ctx)
	s.Require().Equal(types.DefaultParams(), params)

	// Set custom params
	custom := types.Params{Enabled: false, LookbackBlocks: 128}
	err := s.keeper.SetParams(s.ctx, custom)
	s.Require().NoError(err)

	got := s.keeper.GetParams(s.ctx)
	s.Require().Equal(false, got.Enabled)
	s.Require().Equal(uint64(128), got.LookbackBlocks)
}

func (s *KeeperTestSuite) TestSetParams_Invalid() {
	// First set valid params
	err := s.keeper.SetParams(s.ctx, types.DefaultParams())
	s.Require().NoError(err)

	// Try setting invalid params
	invalid := types.Params{Enabled: true, LookbackBlocks: 0}
	err = s.keeper.SetParams(s.ctx, invalid)
	s.Require().Error(err)

	// Verify store was not updated
	got := s.keeper.GetParams(s.ctx)
	s.Require().Equal(types.DefaultParams(), got)
}
