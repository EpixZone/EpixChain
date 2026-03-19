package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/evm/x/xid/types"
)

// NewQueryCmd returns the xid module's root query command
func NewQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "xID identity and DNS query subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		NewResolveNameCmd(),
		NewReverseResolveCmd(),
		NewGetProfileCmd(),
		NewGetDNSRecordsCmd(),
		NewGetTLDCmd(),
		NewListTLDsCmd(),
		NewGetNamesByOwnerCmd(),
		NewParamsCmd(),
		NewGetRegistrationFeeCmd(),
	)
	return cmd
}

// NewResolveNameCmd returns the command for resolving a name
func NewResolveNameCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resolve [name] [tld]",
		Short: "Resolve a name to an owner address (e.g., resolve alice epix)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.ResolveName(cmd.Context(), &types.QueryResolveNameRequest{
				Name: args[0],
				Tld:  args[1],
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// NewReverseResolveCmd returns the command for reverse resolving an address
func NewReverseResolveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reverse-resolve [address]",
		Short: "Find the primary name for an address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.ReverseResolve(cmd.Context(), &types.QueryReverseResolveRequest{
				Address: args[0],
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// NewGetProfileCmd returns the command for getting a profile
func NewGetProfileCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "profile [name] [tld]",
		Short: "Get the profile for a name",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.GetProfile(cmd.Context(), &types.QueryGetProfileRequest{
				Name: args[0],
				Tld:  args[1],
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// NewGetDNSRecordsCmd returns the command for getting DNS records
func NewGetDNSRecordsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dns-records [name] [tld]",
		Short: "Get all DNS records for a name",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.GetDNSRecords(cmd.Context(), &types.QueryGetDNSRecordsRequest{
				Name: args[0],
				Tld:  args[1],
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// NewGetTLDCmd returns the command for getting a TLD config
func NewGetTLDCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tld [name]",
		Short: "Get the configuration for a TLD",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.GetTLD(cmd.Context(), &types.QueryGetTLDRequest{
				Tld: args[0],
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// NewListTLDsCmd returns the command for listing TLDs
func NewListTLDsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tlds",
		Short: "List all registered TLDs",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.ListTLDs(cmd.Context(), &types.QueryListTLDsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// NewGetNamesByOwnerCmd returns the command for listing names by owner
func NewGetNamesByOwnerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "names-by-owner [address]",
		Short: "List all names owned by an address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.GetNamesByOwner(cmd.Context(), &types.QueryGetNamesByOwnerRequest{
				Owner: args[0],
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// NewParamsCmd returns the command for querying module params
func NewParamsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Get the xID module parameters",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.Params(cmd.Context(), &types.QueryParamsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// NewGetRegistrationFeeCmd returns the command for checking registration fees
func NewGetRegistrationFeeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "registration-fee [name] [tld]",
		Short: "Get the registration fee for a name under a TLD",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.GetRegistrationFee(cmd.Context(), &types.QueryGetRegistrationFeeRequest{
				Name: args[0],
				Tld:  args[1],
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
