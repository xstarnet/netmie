#!/bin/bash
# Generate encryption key for vconfig

cd "$(dirname "$0")"

go run -run TestGenerateKey << 'EOF'
package main

import (
	"fmt"
	"crypto/rand"
	"encoding/base64"
	"io"
)

func main() {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	encoded := base64.StdEncoding.EncodeToString(key)
	fmt.Printf("Generated encryption key:\n%s\n\n", encoded)
	fmt.Printf("Update client/cmd/encryption_key.txt with this key\n")
	fmt.Printf("Then use this key on server side to encrypt config:\n")
	fmt.Printf("  go run . encrypt --key '%s' --config config.json\n", encoded)
}
EOF
