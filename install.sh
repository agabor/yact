#!/bin/bash

set -e

echo "Building yact..."
go build -o y

if [ $? -ne 0 ]; then
    echo "Build failed!"
    exit 1
fi

if [ $? -ne 0 ]; then
    echo "Failed to copy system prompts"
    exit 1
fi

echo "Installing y to /usr/local/bin..."
sudo mv y /usr/local/bin/

if [ $? -eq 0 ]; then
    echo "Installation successful!"
    echo "You can now run 'y' from anywhere."
else
    echo "Installation failed. You may need to run this script with sudo."
    exit 1
fi
