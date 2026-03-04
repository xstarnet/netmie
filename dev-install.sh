#!/bin/bash
# Development installation script for Netmie
# This script installs Netmie from local build

set -e

SUDO=""
if command -v sudo > /dev/null && [ "$(id -u)" -ne 0 ]; then
    SUDO="sudo"
fi

echo "Building Netmie..."
go build -o ./netmie ./client

echo "Installing Netmie binary..."
${SUDO} cp ./netmie /usr/local/bin/netmie
${SUDO} chmod +x /usr/local/bin/netmie

echo "Installing Netmie service..."
${SUDO} netmie service install

echo "Starting Netmie service..."
${SUDO} netmie service start

echo ""
echo "Installation complete!"
echo ""
echo "To connect to NetBird network:"
echo "  netmie up"
echo ""
echo "To start V2Ray proxy:"
echo "  netmie vconfig <config-file>"
echo "  netmie vup"
echo ""
echo "Check status:"
echo "  netmie status"
echo "  netmie vstatus"
