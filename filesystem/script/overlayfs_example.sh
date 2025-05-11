#!/bin/bash

# Script to set up overlayfs with a read-only lower layer

# ANSI color codes for better readability
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${GREEN}OverlayFS Setup Script${NC}"
echo "This script will help you set up an overlayfs mount with a read-only lower layer."
echo

# Get user input for paths
read -p "Enter the path to your read-only filesystem (e.g., /mnt/viscaufs): " LOWER_DIR
read -p "Enter the path for the upper (writable) layer: " UPPER_DIR
read -p "Enter the path for the work directory: " WORK_DIR
read -p "Enter the path for the merged view: " MERGED_DIR

# Validate inputs
if [ -z "$LOWER_DIR" ] || [ -z "$UPPER_DIR" ] || [ -z "$WORK_DIR" ] || [ -z "$MERGED_DIR" ]; then
    echo "Error: All paths must be provided."
    exit 1
fi

# Check if lower directory exists
if [ ! -d "$LOWER_DIR" ]; then
    echo "Error: Lower directory '$LOWER_DIR' does not exist."
    exit 1
fi

# Create directories if they don't exist
for DIR in "$UPPER_DIR" "$WORK_DIR" "$MERGED_DIR"; do
    if [ ! -d "$DIR" ]; then
        echo "Creating directory: $DIR"
        sudo mkdir -p "$DIR"
        if [ $? -ne 0 ]; then
            echo "Error: Failed to create directory '$DIR'."
            exit 1
        fi
    fi
done

# Mount overlayfs
echo -e "\n${BLUE}Mounting overlayfs...${NC}"
sudo mount -t overlay overlay \
    -o lowerdir="$LOWER_DIR",upperdir="$UPPER_DIR",workdir="$WORK_DIR" \
    "$MERGED_DIR"

if [ $? -eq 0 ]; then
    echo -e "${GREEN}Success!${NC} OverlayFS mounted successfully."
    echo
    echo "Your setup:"
    echo "  - Read-only lower layer: $LOWER_DIR"
    echo "  - Writable upper layer: $UPPER_DIR"
    echo "  - Work directory: $WORK_DIR"
    echo "  - Merged view: $MERGED_DIR"
    echo
    echo "You can now access and modify files through the merged view at '$MERGED_DIR'."
    echo "Any changes will be stored in the upper layer at '$UPPER_DIR'."
    echo "The original files in '$LOWER_DIR' will remain unchanged."
else
    echo "Error: Failed to mount overlayfs."
    exit 1
fi

echo -e "\n${GREEN}Setup complete!${NC}"
