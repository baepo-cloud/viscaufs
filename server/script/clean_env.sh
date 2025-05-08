#!/bin/bash

# Define the database file
DB_FILE="db/storage.db"

# Check if the database file exists
if [ ! -f "$DB_FILE" ]; then
    echo "Database file $DB_FILE does not exist."
    exit 1
fi

# Check if the -f flag is passed
if [[ "$*" == *"-f"* ]]; then
    # SQL commands to drop tables
    SQL_COMMANDS="
    drop table if exists image_layers;
    drop table if exists images;
    drop table if exists layers;
    drop table if exists schema_migrations;
    "
else
    # SQL commands to truncate tables
    SQL_COMMANDS="
    delete from image_layers;
    delete from images;
    delete from layers;
    "
fi

# Execute SQL commands using sqlite3
echo "Modifying tables in $DB_FILE..."
echo "$SQL_COMMANDS" | sqlite3 "$DB_FILE"

# Check if the images folder exists and remove it
IMAGES_FOLDER="images"
if [ -d "$IMAGES_FOLDER" ]; then
    echo "Removing $IMAGES_FOLDER folder..."
    rm -rf "$IMAGES_FOLDER"
else
    echo "$IMAGES_FOLDER folder does not exist."
fi

echo "Operation completed."
