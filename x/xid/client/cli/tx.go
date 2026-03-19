package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/evm/x/xid/types"
)

// NewTxCmd returns the xid module's root tx command
func NewTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "xID identity and DNS transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		NewRegisterNameCmd(),
		NewTransferNameCmd(),
		NewUpdateProfileCmd(),
		NewSetDNSRecordCmd(),
		NewDeleteDNSRecordCmd(),
	)
	return txCmd
}

// NewRegisterNameCmd returns the command for registering a name
func NewRegisterNameCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register [name] [tld]",
		Short: "Register a name under a TLD (e.g., register alice epix)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgRegisterName{
				Owner: clientCtx.GetFromAddress().String(),
				Name:  args[0],
				Tld:   args[1],
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// NewTransferNameCmd returns the command for transferring a name
func NewTransferNameCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfer [name] [tld] [new-owner]",
		Short: "Transfer ownership of a name to a new address",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgTransferName{
				Owner:    clientCtx.GetFromAddress().String(),
				Name:     args[0],
				Tld:      args[1],
				NewOwner: args[2],
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// NewUpdateProfileCmd returns the command for updating a profile
func NewUpdateProfileCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-profile [name] [tld] [avatar-url] [bio]",
		Short: "Update the profile for a name",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgUpdateProfile{
				Owner: clientCtx.GetFromAddress().String(),
				Name:  args[0],
				Tld:   args[1],
				Profile: types.Profile{
					Avatar: args[2],
					Bio:    args[3],
				},
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// NewSetDNSRecordCmd returns the command for setting a DNS record
func NewSetDNSRecordCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-dns [name] [tld] [record-type] [value] [ttl]",
		Short: "Set a DNS record for a name (record-type: 1=A, 28=AAAA, 5=CNAME, 16=TXT)",
		Args:  cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			var recordType uint32
			if _, err := fmt.Sscan(args[2], &recordType); err != nil {
				return fmt.Errorf("invalid record type: %s", args[2])
			}

			var ttl uint32
			if _, err := fmt.Sscan(args[4], &ttl); err != nil {
				return fmt.Errorf("invalid TTL: %s", args[4])
			}

			msg := &types.MsgSetDNSRecord{
				Owner: clientCtx.GetFromAddress().String(),
				Name:  args[0],
				Tld:   args[1],
				Record: types.DNSRecord{
					RecordType: recordType,
					Value:      args[3],
					Ttl:        ttl,
				},
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// NewDeleteDNSRecordCmd returns the command for deleting a DNS record
func NewDeleteDNSRecordCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete-dns [name] [tld] [record-type]",
		Short: "Delete a DNS record from a name",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			var recordType uint32
			if _, err := fmt.Sscan(args[2], &recordType); err != nil {
				return fmt.Errorf("invalid record type: %s", args[2])
			}

			msg := &types.MsgDeleteDNSRecord{
				Owner:      clientCtx.GetFromAddress().String(),
				Name:       args[0],
				Tld:        args[1],
				RecordType: recordType,
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
