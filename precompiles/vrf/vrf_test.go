package vrf_test

import (
	"crypto/sha256"
	"encoding/hex"
	"math/big"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	vrfprecompile "github.com/cosmos/evm/precompiles/vrf"
	"github.com/cosmos/evm/x/vrf/keeper"
	"github.com/cosmos/evm/x/vrf/types"

	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func setupPrecompile(t *testing.T) (*vrfprecompile.Precompile, keeper.Keeper, testutil.TestContext) {
	t.Helper()
	encCfg := moduletestutil.MakeTestEncodingConfig()
	key := storetypes.NewKVStoreKey(types.StoreKey)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	k := keeper.NewKeeper(encCfg.Codec, key, authtypes.NewModuleAddress("gov").String())
	p := vrfprecompile.NewPrecompile(k)
	return p, k, testCtx
}

func makeTestBeacon(height uint64, beaconHex string) types.RandomBeacon {
	return types.RandomBeacon{
		Height:    height,
		Beacon:    beaconHex,
		Proposer:  "cosmosvalcons1test",
		Timestamp: int64(height) * 1000, //nolint:gosec // G115
	}
}

func TestABIParsing(t *testing.T) {
	p, _, _ := setupPrecompile(t)

	// Should have 3 methods
	require.NotNil(t, p.Methods["getBeacon"])
	require.NotNil(t, p.Methods["latestBeacon"])
	require.NotNil(t, p.Methods["getMultiBlockBeacon"])

	// getBeacon: 1 input (uint64), 1 output (bytes32)
	m := p.Methods["getBeacon"]
	require.Len(t, m.Inputs, 1)
	require.Equal(t, "uint64", m.Inputs[0].Type.String())
	require.Len(t, m.Outputs, 1)
	require.Equal(t, "bytes32", m.Outputs[0].Type.String())

	// latestBeacon: 0 inputs, 2 outputs (bytes32, uint64)
	m = p.Methods["latestBeacon"]
	require.Len(t, m.Inputs, 0)
	require.Len(t, m.Outputs, 2)
	require.Equal(t, "bytes32", m.Outputs[0].Type.String())
	require.Equal(t, "uint64", m.Outputs[1].Type.String())

	// getMultiBlockBeacon: 2 inputs (uint64, uint64), 1 output (bytes32)
	m = p.Methods["getMultiBlockBeacon"]
	require.Len(t, m.Inputs, 2)
	require.Equal(t, "uint64", m.Inputs[0].Type.String())
	require.Equal(t, "uint64", m.Inputs[1].Type.String())
	require.Len(t, m.Outputs, 1)
}

func TestIsTransaction(t *testing.T) {
	p, _, _ := setupPrecompile(t)

	for _, name := range []string{"getBeacon", "latestBeacon", "getMultiBlockBeacon"} {
		m := p.Methods[name]
		require.False(t, p.IsTransaction(&m), "%s should not be a transaction", name)
	}
}

func TestGetBeacon_Found(t *testing.T) {
	p, k, testCtx := setupPrecompile(t)
	ctx := testCtx.Ctx

	beaconHex := strings.Repeat("ab", 32)
	k.SetBeacon(ctx, makeTestBeacon(42, beaconHex))

	method := p.Methods["getBeacon"]
	bz, err := p.GetBeacon(ctx, &method, []interface{}{uint64(42)})
	require.NoError(t, err)

	// Unpack and verify
	results, err := method.Outputs.Unpack(bz)
	require.NoError(t, err)
	require.Len(t, results, 1)

	resultBytes := results[0].([32]byte)
	expectedBytes, _ := hex.DecodeString(beaconHex)
	var expected [32]byte
	copy(expected[:], expectedBytes)
	require.Equal(t, expected, resultBytes)
}

func TestGetBeacon_NotFound(t *testing.T) {
	p, _, testCtx := setupPrecompile(t)
	ctx := testCtx.Ctx

	method := p.Methods["getBeacon"]
	bz, err := p.GetBeacon(ctx, &method, []interface{}{uint64(999)})
	require.NoError(t, err)

	results, err := method.Outputs.Unpack(bz)
	require.NoError(t, err)
	resultBytes := results[0].([32]byte)
	require.Equal(t, [32]byte{}, resultBytes)
}

func TestGetBeacon_BigIntArg(t *testing.T) {
	p, k, testCtx := setupPrecompile(t)
	ctx := testCtx.Ctx

	beaconHex := strings.Repeat("cd", 32)
	k.SetBeacon(ctx, makeTestBeacon(42, beaconHex))

	method := p.Methods["getBeacon"]
	bz, err := p.GetBeacon(ctx, &method, []interface{}{big.NewInt(42)})
	require.NoError(t, err)

	results, err := method.Outputs.Unpack(bz)
	require.NoError(t, err)
	resultBytes := results[0].([32]byte)
	require.NotEqual(t, [32]byte{}, resultBytes) // should find it
}

func TestGetBeacon_InvalidArgs(t *testing.T) {
	p, _, testCtx := setupPrecompile(t)
	ctx := testCtx.Ctx
	method := p.Methods["getBeacon"]

	// Wrong arg count
	_, err := p.GetBeacon(ctx, &method, []interface{}{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "expected 1 argument")

	_, err = p.GetBeacon(ctx, &method, []interface{}{uint64(1), uint64(2)})
	require.Error(t, err)

	// Wrong type
	_, err = p.GetBeacon(ctx, &method, []interface{}{"not a number"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid argument type")
}

func TestLatestBeacon_Found(t *testing.T) {
	p, k, testCtx := setupPrecompile(t)
	ctx := testCtx.Ctx

	beaconHex := strings.Repeat("ef", 32)
	k.SetBeacon(ctx, makeTestBeacon(100, beaconHex))
	k.SetLatestHeight(ctx, 100)

	method := p.Methods["latestBeacon"]
	bz, err := p.LatestBeacon(ctx, &method, []interface{}{})
	require.NoError(t, err)

	results, err := method.Outputs.Unpack(bz)
	require.NoError(t, err)
	require.Len(t, results, 2)

	resultBytes := results[0].([32]byte)
	resultHeight := results[1].(uint64)

	require.NotEqual(t, [32]byte{}, resultBytes)
	require.Equal(t, uint64(100), resultHeight)
}

func TestLatestBeacon_NotFound(t *testing.T) {
	p, _, testCtx := setupPrecompile(t)
	ctx := testCtx.Ctx

	method := p.Methods["latestBeacon"]
	bz, err := p.LatestBeacon(ctx, &method, []interface{}{})
	require.NoError(t, err)

	results, err := method.Outputs.Unpack(bz)
	require.NoError(t, err)
	resultBytes := results[0].([32]byte)
	resultHeight := results[1].(uint64)
	require.Equal(t, [32]byte{}, resultBytes)
	require.Equal(t, uint64(0), resultHeight)
}

func TestGetMultiBlockBeacon_Success(t *testing.T) {
	p, k, testCtx := setupPrecompile(t)
	ctx := testCtx.Ctx

	beacon8 := strings.Repeat("aa", 32)
	beacon9 := strings.Repeat("bb", 32)
	beacon10 := strings.Repeat("cc", 32)

	k.SetBeacon(ctx, makeTestBeacon(8, beacon8))
	k.SetBeacon(ctx, makeTestBeacon(9, beacon9))
	k.SetBeacon(ctx, makeTestBeacon(10, beacon10))

	method := p.Methods["getMultiBlockBeacon"]
	bz, err := p.GetMultiBlockBeacon(ctx, &method, []interface{}{uint64(10), uint64(3)})
	require.NoError(t, err)

	results, err := method.Outputs.Unpack(bz)
	require.NoError(t, err)
	resultBytes := results[0].([32]byte)

	// Manually compute expected: SHA256(beacon8 || beacon9 || beacon10)
	h := sha256.New()
	h.Write([]byte(beacon8))
	h.Write([]byte(beacon9))
	h.Write([]byte(beacon10))
	var expected [32]byte
	copy(expected[:], h.Sum(nil))

	require.Equal(t, expected, resultBytes)
}

func TestGetMultiBlockBeacon_SingleBlock(t *testing.T) {
	p, k, testCtx := setupPrecompile(t)
	ctx := testCtx.Ctx

	beacon10 := strings.Repeat("dd", 32)
	k.SetBeacon(ctx, makeTestBeacon(10, beacon10))

	method := p.Methods["getMultiBlockBeacon"]
	bz, err := p.GetMultiBlockBeacon(ctx, &method, []interface{}{uint64(10), uint64(1)})
	require.NoError(t, err)

	results, err := method.Outputs.Unpack(bz)
	require.NoError(t, err)
	resultBytes := results[0].([32]byte)

	// SHA256 of just one beacon string
	h := sha256.New()
	h.Write([]byte(beacon10))
	var expected [32]byte
	copy(expected[:], h.Sum(nil))

	require.Equal(t, expected, resultBytes)
}

func TestGetMultiBlockBeacon_MissingInRange(t *testing.T) {
	p, k, testCtx := setupPrecompile(t)
	ctx := testCtx.Ctx

	// Store 8 and 10 but NOT 9
	k.SetBeacon(ctx, makeTestBeacon(8, strings.Repeat("aa", 32)))
	k.SetBeacon(ctx, makeTestBeacon(10, strings.Repeat("cc", 32)))

	method := p.Methods["getMultiBlockBeacon"]
	bz, err := p.GetMultiBlockBeacon(ctx, &method, []interface{}{uint64(10), uint64(3)})
	require.NoError(t, err)

	results, err := method.Outputs.Unpack(bz)
	require.NoError(t, err)
	resultBytes := results[0].([32]byte)
	require.Equal(t, [32]byte{}, resultBytes, "should return zero when beacon is missing in range")
}

func TestGetMultiBlockBeacon_InvalidBlocks(t *testing.T) {
	p, _, testCtx := setupPrecompile(t)
	ctx := testCtx.Ctx
	method := p.Methods["getMultiBlockBeacon"]

	// blocks = 0
	_, err := p.GetMultiBlockBeacon(ctx, &method, []interface{}{uint64(10), uint64(0)})
	require.Error(t, err)
	require.Contains(t, err.Error(), "blocks must be between 1 and 256")

	// blocks > 256
	_, err = p.GetMultiBlockBeacon(ctx, &method, []interface{}{uint64(300), uint64(257)})
	require.Error(t, err)
	require.Contains(t, err.Error(), "blocks must be between 1 and 256")
}

func TestGetMultiBlockBeacon_EndHeightLessThanBlocks(t *testing.T) {
	p, _, testCtx := setupPrecompile(t)
	ctx := testCtx.Ctx
	method := p.Methods["getMultiBlockBeacon"]

	_, err := p.GetMultiBlockBeacon(ctx, &method, []interface{}{uint64(2), uint64(5)})
	require.Error(t, err)
	require.Contains(t, err.Error(), "endHeight must be >= blocks")
}

func TestGetMultiBlockBeacon_BigIntArgs(t *testing.T) {
	p, k, testCtx := setupPrecompile(t)
	ctx := testCtx.Ctx

	beacon10 := strings.Repeat("ee", 32)
	k.SetBeacon(ctx, makeTestBeacon(10, beacon10))

	method := p.Methods["getMultiBlockBeacon"]
	bz, err := p.GetMultiBlockBeacon(ctx, &method, []interface{}{big.NewInt(10), big.NewInt(1)})
	require.NoError(t, err)

	results, err := method.Outputs.Unpack(bz)
	require.NoError(t, err)
	resultBytes := results[0].([32]byte)
	require.NotEqual(t, [32]byte{}, resultBytes)
}

func TestHexToBytes32(t *testing.T) {
	// Valid 64-char hex
	validHex := strings.Repeat("ab", 32)
	p, _, _ := setupPrecompile(t)
	_ = p // just to ensure setup works

	decoded, _ := hex.DecodeString(validHex)
	require.Len(t, decoded, 32)
}

func TestRequiredGas(t *testing.T) {
	p, _, _ := setupPrecompile(t)

	// Input too short
	require.Equal(t, uint64(0), p.RequiredGas([]byte{1, 2, 3}))

	// Valid method ID for getBeacon
	input, err := p.Pack("getBeacon", uint64(42))
	require.NoError(t, err)
	gas := p.RequiredGas(input)
	require.Greater(t, gas, uint64(0))

	// Valid method ID for latestBeacon
	input, err = p.Pack("latestBeacon")
	require.NoError(t, err)
	gas = p.RequiredGas(input)
	require.Greater(t, gas, uint64(0))

	// Valid method ID for getMultiBlockBeacon
	input, err = p.Pack("getMultiBlockBeacon", uint64(10), uint64(3))
	require.NoError(t, err)
	gas = p.RequiredGas(input)
	require.Greater(t, gas, uint64(0))

	// Unknown method ID
	require.Equal(t, uint64(0), p.RequiredGas([]byte{0xFF, 0xFF, 0xFF, 0xFF}))
}
