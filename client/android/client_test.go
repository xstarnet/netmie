//go:build android

package android

import (
	"testing"

	"github.com/netbirdio/netbird/client/internal"
	"github.com/stretchr/testify/assert"
)

func TestConnectionTypeInitialization(t *testing.T) {
	client := &Client{}
	assert.Equal(t, internal.ConnectionTypeNone, client.currentConnection)
}

func TestConnectionTypeString(t *testing.T) {
	assert.Equal(t, "None", internal.ConnectionTypeNone.String())
	assert.Equal(t, "NetBird", internal.ConnectionTypeNetBird.String())
	assert.Equal(t, "V2Ray", internal.ConnectionTypeV2Ray.String())
}

func TestConnectNetBirdIdempotent(t *testing.T) {
	client := &Client{}
	client.currentConnection = internal.ConnectionTypeNetBird

	err := client.ConnectNetBird()
	assert.NoError(t, err)
	assert.Equal(t, internal.ConnectionTypeNetBird, client.currentConnection)
}

func TestConnectV2RayIdempotent(t *testing.T) {
	client := &Client{}
	client.currentConnection = internal.ConnectionTypeV2Ray

	err := client.ConnectV2Ray("/path/to/config.json")
	assert.NoError(t, err)
	assert.Equal(t, internal.ConnectionTypeV2Ray, client.currentConnection)
}

func TestDisconnectWhenNone(t *testing.T) {
	client := &Client{}
	client.currentConnection = internal.ConnectionTypeNone

	err := client.Disconnect()
	assert.NoError(t, err)
	assert.Equal(t, internal.ConnectionTypeNone, client.currentConnection)
}
