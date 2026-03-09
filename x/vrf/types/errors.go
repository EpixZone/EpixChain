package types

import "cosmossdk.io/errors"

var (
	ErrBeaconNotFound = errors.Register(ModuleName, 2, "beacon not found for height")
	ErrBeaconDisabled = errors.Register(ModuleName, 3, "VRF beacon is disabled")
)
