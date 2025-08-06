#!/bin/bash

# Serve Docsify documentation locally
# This script provides multiple options for serving the documentation

echo "🚀 Starting Docsify documentation server..."

# Check if docsify-cli is installed
if command -v docsify &> /dev/null; then
    echo "📖 Using docsify-cli to serve documentation"
    echo "🌐 Documentation will be available at: http://localhost:3000"
    docsify serve . --port 3000
elif command -v python3 &> /dev/null; then
    echo "🐍 Using Python 3 to serve documentation"
    echo "🌐 Documentation will be available at: http://localhost:3000"
    python3 -m http.server 3000
elif command -v python &> /dev/null; then
    echo "🐍 Using Python to serve documentation"
    echo "🌐 Documentation will be available at: http://localhost:3000"
    python -m http.server 3000
elif command -v node &> /dev/null; then
    echo "📦 Installing docsify-cli and serving documentation"
    npx docsify-cli serve . --port 3000
else
    echo "❌ No suitable server found. Please install one of the following:"
    echo "   - docsify-cli: npm install -g docsify-cli"
    echo "   - Python 3: python3 -m http.server 3000"
    echo "   - Node.js: npx docsify-cli serve . --port 3000"
    exit 1
fi