package xid

import (
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"

	cmn "github.com/cosmos/evm/precompiles/common"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// Event names matching abi.json
	EventNameRegistered     = "NameRegistered"
	EventNameTransferred    = "NameTransferred"
	EventProfileUpdated     = "ProfileUpdated"
	EventDNSRecordSet       = "DNSRecordSet"
	EventDNSRecordDeleted   = "DNSRecordDeleted"
	EventIdentityLinked     = "IdentityLinked"
	EventIdentityUnlinked   = "IdentityUnlinked"
	EventContentRootUpdated = "ContentRootUpdated"
	EventPrimaryNameSet     = "PrimaryNameSet"
)

// EmitNameRegistered emits a NameRegistered event to the EVM state DB.
func (p Precompile) EmitNameRegistered(
	ctx sdk.Context,
	stateDB vm.StateDB,
	owner common.Address,
	name, tld string,
) error {
	event := p.Events[EventNameRegistered]

	// Topic[0] = event signature hash
	// Topic[1] = indexed owner address
	topics := make([]common.Hash, 2)
	topics[0] = event.ID

	ownerTopic, err := cmn.MakeTopic(owner)
	if err != nil {
		return err
	}
	topics[1] = ownerTopic

	// Non-indexed arguments: name, tld
	arguments := event.Inputs.NonIndexed()
	packed, err := arguments.Pack(name, tld)
	if err != nil {
		return err
	}

	stateDB.AddLog(&ethtypes.Log{
		Address:     p.Address(),
		Topics:      topics,
		Data:        packed,
		BlockNumber: uint64(ctx.BlockHeight()),
	})

	return nil
}

// EmitNameTransferred emits a NameTransferred event to the EVM state DB.
func (p Precompile) EmitNameTransferred(
	ctx sdk.Context,
	stateDB vm.StateDB,
	from, to common.Address,
	name, tld string,
) error {
	event := p.Events[EventNameTransferred]

	topics := make([]common.Hash, 3)
	topics[0] = event.ID

	fromTopic, err := cmn.MakeTopic(from)
	if err != nil {
		return err
	}
	topics[1] = fromTopic

	toTopic, err := cmn.MakeTopic(to)
	if err != nil {
		return err
	}
	topics[2] = toTopic

	arguments := event.Inputs.NonIndexed()
	packed, err := arguments.Pack(name, tld)
	if err != nil {
		return err
	}

	stateDB.AddLog(&ethtypes.Log{
		Address:     p.Address(),
		Topics:      topics,
		Data:        packed,
		BlockNumber: uint64(ctx.BlockHeight()),
	})

	return nil
}

// EmitProfileUpdated emits a ProfileUpdated event to the EVM state DB.
func (p Precompile) EmitProfileUpdated(
	ctx sdk.Context,
	stateDB vm.StateDB,
	owner common.Address,
	name, tld string,
) error {
	event := p.Events[EventProfileUpdated]

	topics := make([]common.Hash, 2)
	topics[0] = event.ID

	ownerTopic, err := cmn.MakeTopic(owner)
	if err != nil {
		return err
	}
	topics[1] = ownerTopic

	arguments := event.Inputs.NonIndexed()
	packed, err := arguments.Pack(name, tld)
	if err != nil {
		return err
	}

	stateDB.AddLog(&ethtypes.Log{
		Address:     p.Address(),
		Topics:      topics,
		Data:        packed,
		BlockNumber: uint64(ctx.BlockHeight()),
	})

	return nil
}

// EmitDNSRecordSet emits a DNSRecordSet event to the EVM state DB.
func (p Precompile) EmitDNSRecordSet(
	ctx sdk.Context,
	stateDB vm.StateDB,
	name, tld string,
	recordType uint16,
	value string,
) error {
	event := p.Events[EventDNSRecordSet]

	topics := make([]common.Hash, 1)
	topics[0] = event.ID

	arguments := event.Inputs.NonIndexed()
	packed, err := arguments.Pack(name, tld, recordType, value)
	if err != nil {
		return err
	}

	stateDB.AddLog(&ethtypes.Log{
		Address:     p.Address(),
		Topics:      topics,
		Data:        packed,
		BlockNumber: uint64(ctx.BlockHeight()),
	})

	return nil
}

// EmitDNSRecordDeleted emits a DNSRecordDeleted event to the EVM state DB.
func (p Precompile) EmitDNSRecordDeleted(
	ctx sdk.Context,
	stateDB vm.StateDB,
	name, tld string,
	recordType uint16,
) error {
	event := p.Events[EventDNSRecordDeleted]

	topics := make([]common.Hash, 1)
	topics[0] = event.ID

	arguments := event.Inputs.NonIndexed()
	packed, err := arguments.Pack(name, tld, recordType)
	if err != nil {
		return err
	}

	stateDB.AddLog(&ethtypes.Log{
		Address:     p.Address(),
		Topics:      topics,
		Data:        packed,
		BlockNumber: uint64(ctx.BlockHeight()),
	})

	return nil
}

// EmitIdentityLinked emits an IdentityLinked event to the EVM state DB.
func (p Precompile) EmitIdentityLinked(
	ctx sdk.Context,
	stateDB vm.StateDB,
	name, tld, identityAddress, label string,
) error {
	event := p.Events[EventIdentityLinked]

	topics := make([]common.Hash, 1)
	topics[0] = event.ID

	arguments := event.Inputs.NonIndexed()
	packed, err := arguments.Pack(name, tld, identityAddress, label)
	if err != nil {
		return err
	}

	stateDB.AddLog(&ethtypes.Log{
		Address:     p.Address(),
		Topics:      topics,
		Data:        packed,
		BlockNumber: uint64(ctx.BlockHeight()),
	})

	return nil
}

// EmitIdentityUnlinked emits an IdentityUnlinked event to the EVM state DB.
func (p Precompile) EmitIdentityUnlinked(
	ctx sdk.Context,
	stateDB vm.StateDB,
	name, tld, identityAddress string,
) error {
	event := p.Events[EventIdentityUnlinked]

	topics := make([]common.Hash, 1)
	topics[0] = event.ID

	arguments := event.Inputs.NonIndexed()
	packed, err := arguments.Pack(name, tld, identityAddress)
	if err != nil {
		return err
	}

	stateDB.AddLog(&ethtypes.Log{
		Address:     p.Address(),
		Topics:      topics,
		Data:        packed,
		BlockNumber: uint64(ctx.BlockHeight()),
	})

	return nil
}

// EmitPrimaryNameSet emits a PrimaryNameSet event to the EVM state DB.
func (p Precompile) EmitPrimaryNameSet(
	ctx sdk.Context,
	stateDB vm.StateDB,
	owner common.Address,
	name, tld string,
) error {
	event := p.Events[EventPrimaryNameSet]

	topics := make([]common.Hash, 2)
	topics[0] = event.ID

	ownerTopic, err := cmn.MakeTopic(owner)
	if err != nil {
		return err
	}
	topics[1] = ownerTopic

	arguments := event.Inputs.NonIndexed()
	packed, err := arguments.Pack(name, tld)
	if err != nil {
		return err
	}

	stateDB.AddLog(&ethtypes.Log{
		Address:     p.Address(),
		Topics:      topics,
		Data:        packed,
		BlockNumber: uint64(ctx.BlockHeight()),
	})

	return nil
}

// EmitContentRootUpdated emits a ContentRootUpdated event to the EVM state DB.
func (p Precompile) EmitContentRootUpdated(
	ctx sdk.Context,
	stateDB vm.StateDB,
	name, tld, root string,
) error {
	event := p.Events[EventContentRootUpdated]

	topics := make([]common.Hash, 1)
	topics[0] = event.ID

	arguments := event.Inputs.NonIndexed()
	packed, err := arguments.Pack(name, tld, root)
	if err != nil {
		return err
	}

	stateDB.AddLog(&ethtypes.Log{
		Address:     p.Address(),
		Topics:      topics,
		Data:        packed,
		BlockNumber: uint64(ctx.BlockHeight()),
	})

	return nil
}
