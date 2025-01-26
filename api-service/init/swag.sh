#!/bin/bash

# Exit script on error
set -e

# Export GOPATH and update PATH
echo "Setting up Go environment..."
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin

# Source ~/.bashrc to ensure the environment is updated
echo "Sourcing ~/.bashrc..."
source ~/.bashrc

# Install Swag
echo "Installing Swag CLI..."
go install github.com/swaggo/swag/cmd/swag@latest

# Confirm installation
echo "Swag CLI installed successfully!"