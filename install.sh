#!/bin/bash

set -e

echo "Building yact..."
go build -o y

if [ $? -ne 0 ]; then
    echo "Build failed!"
    exit 1
fi

CONFIG_DIR="$HOME/.yact"
PROMPTS_DIR="$CONFIG_DIR/systemprompts"

mkdir -p "$PROMPTS_DIR"

echo "Copying system prompts to $PROMPTS_DIR..."
cp systemprompts/*.txt "$PROMPTS_DIR/"

if [ $? -ne 0 ]; then
    echo "Failed to copy system prompts"
    exit 1
fi

echo "Installing y to /usr/local/bin..."
sudo mv y /usr/local/bin/

if [ $? -eq 0 ]; then
    echo "Installation successful!"
    echo "System prompts installed to $PROMPTS_DIR"
    echo "You can now run 'y' from anywhere."
else
    echo "Installation failed. You may need to run this script with sudo."
    exit 1
fi