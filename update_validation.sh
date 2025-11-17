#!/bin/bash

# Script to replace validation imports and calls

cd /home/user/message-gateway/handler

# Replace validation imports in all handler files
for file in templates.go providers.go reports.go bulksms.go utility.go; do
    echo "Processing $file..."

    # Add models import if not already present
    if ! grep -q '"MgApplication/models"' "$file"; then
        sed -i '/import (/a\	"MgApplication/models"' "$file"
    fi

    # Remove validation import
    sed -i '/validation "MgApplication\/api-validation"/d' "$file"

    # Replace validation.ValidateStruct calls with .Validate()
    sed -i 's/validation\.ValidateStruct(\([^)]*\))/\1.Validate()/g' "$file"
done

echo "Done!"
