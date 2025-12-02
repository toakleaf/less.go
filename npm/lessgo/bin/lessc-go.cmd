@echo off
setlocal enabledelayedexpansion

:: Wrapper script for lessc-go binary (Windows)
:: This script finds and executes the platform-specific binary

:: Get the directory where this script lives
set "SCRIPT_DIR=%~dp0"
set "PACKAGE_DIR=%SCRIPT_DIR%.."

:: Detect architecture
if "%PROCESSOR_ARCHITECTURE%"=="AMD64" (
    set "ARCH_NAME=x64"
) else if "%PROCESSOR_ARCHITECTURE%"=="ARM64" (
    set "ARCH_NAME=arm64"
) else if "%PROCESSOR_ARCHITEW6432%"=="AMD64" (
    set "ARCH_NAME=x64"
) else (
    echo lessc-go: Unsupported architecture: %PROCESSOR_ARCHITECTURE% 1>&2
    exit /b 1
)

set "PLATFORM_KEY=win32-%ARCH_NAME%"
set "PACKAGE_NAME=@lessgo/%PLATFORM_KEY%"

:: Try to find the binary in node_modules
set "BINARY_PATH="

:: Check in the same node_modules as this package
if exist "%PACKAGE_DIR%\..\%PACKAGE_NAME%\bin\lessc-go.exe" (
    set "BINARY_PATH=%PACKAGE_DIR%\..\%PACKAGE_NAME%\bin\lessc-go.exe"
    goto :run
)

:: Check in node_modules relative to cwd
if exist ".\node_modules\%PACKAGE_NAME%\bin\lessc-go.exe" (
    set "BINARY_PATH=.\node_modules\%PACKAGE_NAME%\bin\lessc-go.exe"
    goto :run
)

:: Use Node.js to resolve
for /f "delims=" %%i in ('node -e "try { const path = require('path'); const pkg = require.resolve('%PACKAGE_NAME%/package.json'); const bin = path.join(path.dirname(pkg), 'bin', 'lessc-go.exe'); console.log(bin); } catch (e) { process.exit(1); }" 2^>nul') do set "BINARY_PATH=%%i"

if not defined BINARY_PATH (
    echo lessc-go: Could not find the platform-specific binary. 1>&2
    echo lessc-go: Platform: %PLATFORM_KEY% 1>&2
    echo lessc-go: Expected package: %PACKAGE_NAME% 1>&2
    echo lessc-go: Try reinstalling lessgo 1>&2
    exit /b 1
)

if not exist "%BINARY_PATH%" (
    echo lessc-go: Binary not found at %BINARY_PATH% 1>&2
    exit /b 1
)

:run
:: Execute the binary with all arguments
"%BINARY_PATH%" %*
exit /b %errorlevel%
