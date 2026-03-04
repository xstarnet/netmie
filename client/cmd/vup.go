package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/netbirdio/netbird/client/internal/v2ray"
	"github.com/netbirdio/netbird/client/proto"
)

var vupCmd = &cobra.Command{
	Use:   "vup",
	Short: "Start V2Ray connection",
	Long:  "Start V2Ray VPN connection using the configured settings",
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

		// Call VUp RPC
		resp, err := client.VUp(cmd.Context(), &proto.VUpRequest{})
		if err != nil {
			return fmt.Errorf("failed to start V2Ray: %v", err)
		}

		if !resp.Success {
			return fmt.Errorf("%s", resp.Message)
		}

		cmd.Println("✓ V2Ray started successfully")

		// Get status info to show proxy endpoints
		statusResp, err := client.VStatus(cmd.Context(), &proto.VStatusRequest{})
		if err == nil && statusResp.Status == "running" {
			if inbounds := getInboundsInfo(); inbounds != "" {
				cmd.Println("\nProxy endpoints:")
				cmd.Print(inbounds)
			}
		}

		return nil
	},
}

func getInboundsInfo() string {
	configPath := v2ray.DefaultConfigPath()
	cm := v2ray.NewConfigManager(configPath)
	if err := cm.Load(); err != nil {
		return ""
	}

	config := cm.Get()
	if config == nil || len(config.Inbounds) == 0 {
		return ""
	}

	result := ""
	for _, inbound := range config.Inbounds {
		listen := inbound.Listen
		if listen == "" || listen == "0.0.0.0" || listen == "::" {
			listen = "127.0.0.1"
		}
		result += fmt.Sprintf("  %s: %s:%d\n", inbound.Protocol, listen, inbound.Port)
	}
	return result
}

func init() {
	rootCmd.AddCommand(vupCmd)
}
