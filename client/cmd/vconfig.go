package cmd

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/spf13/cobra"

	"github.com/netbirdio/netbird/client/proto"
	"github.com/netbirdio/netbird/util/crypt"
)

//go:embed encryption_key.txt
// encryption_key.txt contains the base64-encoded AES-256 encryption key.
// This file is NOT committed to git (see .gitignore).
// To set up: copy encryption_key.txt.example to encryption_key.txt and update with your key.
var encryptionKey string

var (
	vconfigHost string
	vconfigPort int
)

var vconfigCmd = &cobra.Command{
	Use:   "vconfig <config-url>",
	Short: "Update V2Ray configuration from HTTP",
	Long:  "Download and decrypt V2Ray configuration from HTTP endpoint. The configuration must be encrypted with AES-256-GCM.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SetOut(cmd.OutOrStdout())
		configURL := args[0]

		// Download encrypted config from HTTP
		encryptedData, err := downloadConfig(configURL)
		if err != nil {
			return fmt.Errorf("failed to download config: %w", err)
		}

		// Decrypt config with embedded encryption key
		plaintext, err := decryptConfig(encryptedData)
		if err != nil {
			return fmt.Errorf("failed to decrypt config: %w", err)
		}

		// Merge inbounds with server-provided outbounds
		fullConfig, err := mergeInbounds(plaintext, vconfigHost, vconfigPort)
		if err != nil {
			return fmt.Errorf("failed to merge inbounds: %w", err)
		}

		// Write full config to temporary file
		tempFile, err := os.CreateTemp("", "netmie-config-*.json")
		if err != nil {
			return fmt.Errorf("failed to create temp file: %w", err)
		}
		defer os.Remove(tempFile.Name())

		if _, err := tempFile.Write(fullConfig); err != nil {
			tempFile.Close()
			return fmt.Errorf("failed to write config: %w", err)
		}
		tempFile.Close()

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
			ConfigPath: tempFile.Name(),
		})
		if err != nil {
			return fmt.Errorf("failed to update config: %w", err)
		}

		if !resp.Success {
			return fmt.Errorf("%s", resp.Message)
		}

		cmd.Println("✓ V2Ray configuration updated successfully")
		return nil
	},
}

// downloadConfig downloads encrypted config from HTTP endpoint
func downloadConfig(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Parse JSON response
	var response struct {
		Data struct {
			Data      string `json:"data"`
			Encrypted bool   `json:"encrypted"`
		} `json:"data"`
		Code int `json:"code"`
	}

	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if response.Code != 20000 {
		return nil, fmt.Errorf("server returned error code: %d", response.Code)
	}

	if !response.Data.Encrypted {
		return nil, fmt.Errorf("server returned unencrypted data")
	}

	return []byte(response.Data.Data), nil
}

// decryptConfig decrypts AES-256-GCM encrypted config with embedded key
func decryptConfig(encryptedData []byte) ([]byte, error) {
	// Create cipher with embedded key
	cipher, err := crypt.NewFieldEncrypt(encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize cipher: %w", err)
	}

	// Encrypted data is base64-encoded, decode it first
	encryptedB64 := string(encryptedData)
	decrypted, err := cipher.Decrypt(encryptedB64)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	return []byte(decrypted), nil
}

// mergeInbounds merges server-provided outbounds with client-generated inbounds
func mergeInbounds(serverConfig []byte, host string, port int) ([]byte, error) {
	// Parse server config (contains outbounds only)
	var config map[string]interface{}
	if err := json.Unmarshal(serverConfig, &config); err != nil {
		return nil, fmt.Errorf("invalid server config JSON: %w", err)
	}

	// Add inbounds configuration
	config["inbounds"] = []map[string]interface{}{
		{
			"listen":   host,
			"port":     port,
			"protocol": "socks",
			"settings": map[string]interface{}{
				"auth": "noauth",
				"udp":  true,
			},
			"tag": "socks-in",
		},
		{
			"listen":   host,
			"port":     port + 1,
			"protocol": "http",
			"tag":      "http-in",
		},
	}

	// Marshal back to JSON
	fullConfig, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	return fullConfig, nil
}

func init() {
	rootCmd.AddCommand(vconfigCmd)
	vconfigCmd.Flags().StringVar(&vconfigHost, "host", "127.0.0.1", "Inbound listen host")
	vconfigCmd.Flags().IntVar(&vconfigPort, "port", 10808, "Inbound SOCKS5 port (HTTP will be port+1)")
}
