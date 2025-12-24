#!/bin/bash

# Migration script to move files from app/ to src/

echo "Starting migration..."

# Create src directory structure
mkdir -p src/{components,pages,composables,utils,types,config}

# Move directories
if [ -d "app/components" ]; then
    echo "Moving components..."
    cp -r app/components/* src/components/
fi

if [ -d "app/pages" ]; then
    echo "Moving pages..."
    cp -r app/pages/* src/pages/
fi

if [ -d "app/composables" ]; then
    echo "Moving composables..."
    cp -r app/composables/* src/composables/
fi

if [ -d "app/utils" ]; then
    echo "Moving utils..."
    cp -r app/utils/* src/utils/
fi

if [ -d "app/types" ]; then
    echo "Moving types..."
    cp -r app/types/* src/types/
fi

echo "Migration complete!"
echo "Next steps:"
echo "1. Review and update components to replace NuxtLink with RouterLink"
echo "2. Update any remaining Nuxt-specific code"
echo "3. Run: npm install"
echo "4. Run: npm run dev"

