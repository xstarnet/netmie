package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/netbirdio/netbird/client/proto"
)

var vconfigCmd = &cobra.Command{
	Use:   "vconfig <config-file>",
	Short: "Update V2Ray configuration",
	Long:  "Update V2Ray configuration file and restart if running",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SetOut(cmd.OutOrStdout())

		configFile := args[0]

		// Convert to absolute path
		absPath, err := filepath.Abs(configFile)
		if err != nil {
			return fmt.Errorf("failed to get absolute path: %v", err)
		}

		// Check if config file exists
		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			return fmt.Errorf("config file not found: %s", absPath)
		}

		// Connect to daemon
		conn, err := DialClientGRPCServer(cmd.Context(), daemonAddr)
		if err != nil {
			return fmt.Errorf("failed to connect to daemon: %v\n"+
				"Make sure the daemon is running:\n"+
				"  netmie service start", err)
		}
		defer conn.Close()

		client := proto.NewDaemonServiceClient(conn)

		// Call VConfig RPC
		resp, err := client.VConfig(cmd.Context(), &proto.VConfigRequest{
			ConfigPath: absPath,
		})
		if err != nil {
			return fmt.Errorf("failed to update config: %v", err)
		}

		if !resp.Success {
			return fmt.Errorf("%s", resp.Message)
		}

		cmd.Println("✓", resp.Message)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(vconfigCmd)
}
