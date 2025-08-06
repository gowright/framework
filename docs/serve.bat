@echo off
REM Serve Docsify documentation locally on Windows
REM This script provides multiple options for serving the documentation

echo 🚀 Starting Docsify documentation server...

REM Check if docsify-cli is installed
where docsify >nul 2>nul
if %ERRORLEVEL% == 0 (
    echo 📖 Using docsify-cli to serve documentation
    echo 🌐 Documentation will be available at: http://localhost:3000
    docsify serve . --port 3000
    goto :end
)

REM Check if Python is installed
where python >nul 2>nul
if %ERRORLEVEL% == 0 (
    echo 🐍 Using Python to serve documentation
    echo 🌐 Documentation will be available at: http://localhost:3000
    python -m http.server 3000
    goto :end
)

REM Check if Node.js is installed
where node >nul 2>nul
if %ERRORLEVEL% == 0 (
    echo 📦 Installing docsify-cli and serving documentation
    npx docsify-cli serve . --port 3000
    goto :end
)

echo ❌ No suitable server found. Please install one of the following:
echo    - docsify-cli: npm install -g docsify-cli
echo    - Python: python -m http.server 3000
echo    - Node.js: npx docsify-cli serve . --port 3000

:end
pause