package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/netbirdio/netbird/client/proto"
)

var vstatusCmd = &cobra.Command{
	Use:   "vstatus",
	Short: "Show V2Ray connection status",
	Long:  "Display the current status of V2Ray VPN connection",
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

		// Call VStatus RPC
		resp, err := client.VStatus(cmd.Context(), &proto.VStatusRequest{})
		if err != nil {
			return fmt.Errorf("failed to get V2Ray status: %v", err)
		}

		// Display status
		if resp.Status == "running" {
			cmd.Println("V2Ray Status: Running")
			if resp.Interface != "" {
				cmd.Printf("  Interface: %s\n", resp.Interface)
			}
			if resp.Ip != "" {
				cmd.Printf("  IP: %s\n", resp.Ip)
			}

			// Try to get proxy endpoints info
			if inbounds := getInboundsInfo(); inbounds != "" {
				cmd.Println("\nProxy endpoints:")
				cmd.Print(inbounds)
			}
		} else {
			cmd.Println("V2Ray Status: Stopped")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(vstatusCmd)
}
