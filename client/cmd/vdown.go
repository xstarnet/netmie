package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/netbirdio/netbird/client/proto"
)

var vdownCmd = &cobra.Command{
	Use:   "vdown",
	Short: "Stop V2Ray connection",
	Long:  "Stop the running V2Ray VPN connection",
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SetOut(cmd.OutOrStdout())

		// Connect to daemon
		conn, err := DialClientGRPCServer(cmd.Context(), daemonAddr)
		if err != nil {
			return fmt.Errorf("failed to connect to daemon: %v\n"+
				"Make sure the daemon is running:\n"+
				"  netmie service start", err)
		}
		defer conn.Close()

		client := proto.NewDaemonServiceClient(conn)

		// Call VDown RPC
		resp, err := client.VDown(cmd.Context(), &proto.VDownRequest{})
		if err != nil {
			return fmt.Errorf("failed to stop V2Ray: %v", err)
		}

		if !resp.Success {
			return fmt.Errorf("%s", resp.Message)
		}

		cmd.Println("✓ V2Ray stopped successfully")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(vdownCmd)
}
