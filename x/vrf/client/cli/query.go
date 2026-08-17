package cli

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/cosmos/evm/x/vrf/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
)

// NewQueryCmd returns the CLI query commands for the vrf module
func NewQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the VRF module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdGetBeacon(),
		CmdLatestBeacon(),
		CmdQueryParams(),
	)

	return cmd
}

// CmdGetBeacon returns the command to query a beacon by height
func CmdGetBeacon() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-beacon [height]",
		Short: "Query the random beacon at a specific block height",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			height, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid height: %w", err)
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.GetBeacon(cmd.Context(), &types.QueryGetBeaconRequest{Height: height})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res.Beacon)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdLatestBeacon returns the command to query the latest beacon
func CmdLatestBeacon() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "latest-beacon",
		Short: "Query the most recent random beacon",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.LatestBeacon(cmd.Context(), &types.QueryLatestBeaconRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res.Beacon)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryParams returns the command to query module params
func CmdQueryParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Query the VRF module parameters",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.Params(cmd.Context(), &types.QueryParamsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(&res.Params)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
