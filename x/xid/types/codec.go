package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

var (
	amino = codec.NewLegacyAmino()

	// AminoCdc references the global x/xid module codec.
	AminoCdc = codec.NewAminoCodec(amino)
)

// RegisterLegacyAminoCodec registers the necessary x/xid interfaces and concrete types
// on the provided LegacyAmino codec.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgRegisterName{}, "cosmos/evm/x/xid/MsgRegisterName", nil)
	cdc.RegisterConcrete(&MsgTransferName{}, "cosmos/evm/x/xid/MsgTransferName", nil)
	cdc.RegisterConcrete(&MsgUpdateProfile{}, "cosmos/evm/x/xid/MsgUpdateProfile", nil)
	cdc.RegisterConcrete(&MsgSetDNSRecord{}, "cosmos/evm/x/xid/MsgSetDNSRecord", nil)
	cdc.RegisterConcrete(&MsgDeleteDNSRecord{}, "cosmos/evm/x/xid/MsgDeleteDNSRecord", nil)
	cdc.RegisterConcrete(&MsgCreateTLD{}, "cosmos/evm/x/xid/MsgCreateTLD", nil)
	cdc.RegisterConcrete(&MsgUpdateTLDConfig{}, "cosmos/evm/x/xid/MsgUpdateTLDConfig", nil)
	cdc.RegisterConcrete(&MsgUpdateParams{}, "cosmos/evm/x/xid/MsgUpdateParams", nil)
}

// RegisterInterfaces registers the xid module interface types
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgRegisterName{},
		&MsgTransferName{},
		&MsgUpdateProfile{},
		&MsgSetDNSRecord{},
		&MsgDeleteDNSRecord{},
		&MsgCreateTLD{},
		&MsgUpdateTLDConfig{},
		&MsgUpdateParams{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

func init() {
	RegisterLegacyAminoCodec(amino)
	amino.Seal()
}
