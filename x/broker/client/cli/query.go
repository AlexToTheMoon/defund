package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/defund-labs/defund/x/broker/types"
	"github.com/spf13/cobra"
)

// GetQueryCmd creates and returns the broker query command
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the inter-tx module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(getInterchainAccountCmd())
	cmd.AddCommand(getBrokerCmd())

	return cmd
}

func getInterchainAccountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "interchainaccounts [connection-id] [owner-account]",
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.InterchainAccountFromAddress(cmd.Context(), types.NewQueryInterchainAccountRequest(args[0], args[1]))
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func getBrokerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "broker [brokerId]",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.Broker(cmd.Context(), types.NewQueryBrokerRequest(args[0]))
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
