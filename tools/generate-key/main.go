package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

func main() {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	encoded := base64.StdEncoding.EncodeToString(key)
	fmt.Printf("Generated encryption key:\n%s\n", encoded)
}
