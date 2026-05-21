@echo off
setlocal enabledelayedexpansion

REM GCM (GitHub Config Manager) installation script for Windows Command Prompt
REM This script installs gcm to %USERPROFILE%\.local\bin and adds it to PATH

REM Parse command line arguments
set QUIET_MODE=0
set SPECIFIC_VERSION=
set SHOW_HELP=0

:parse_args
if "%~1"=="" goto :args_done
if /i "%~1"=="--quiet" set QUIET_MODE=1 & shift & goto :parse_args
if /i "%~1"=="-q" set QUIET_MODE=1 & shift & goto :parse_args
if /i "%~1"=="--version" set SPECIFIC_VERSION=%~2 & shift & shift & goto :parse_args
if /i "%~1"=="-v" set SPECIFIC_VERSION=%~2 & shift & shift & goto :parse_args
if /i "%~1"=="--help" set SHOW_HELP=1 & shift & goto :parse_args
if /i "%~1"=="-h" set SHOW_HELP=1 & shift & goto :parse_args
echo Unknown option: %~1
call :show_help
exit /b 1

:args_done

REM Show help if requested
if %SHOW_HELP%==1 (
    call :show_help
    exit /b 0
)

REM ANSI color codes (for Windows 10+ terminals)
set "RED=[0;31m"
set "GREEN=[0;32m"
set "YELLOW=[1;33m"
set "BLUE=[0;34m"
set "PURPLE=[0;35m"
set "CYAN=[0;36m"
set "WHITE=[1;37m"
set "GRAY=[0;90m"
set "RESET=[0m"
set "BOLD=[1m"
set "DIM=[2m"

REM Check if we're in a terminal that supports ANSI colors
ver | find "Version 10." >nul
if %errorlevel% neq 0 (
    set "RED="
    set "GREEN="
    set "YELLOW="
    set "BLUE="
    set "PURPLE="
    set "CYAN="
    set "WHITE="
    set "GRAY="
    set "RESET="
    set "BOLD="
    set "DIM="
)

REM Unicode characters (will fallback to ASCII on older systems)
set "CHECKMARK=v"
set "CROSSMARK=x"
set "ARROW=->"
set "INFO=i"
set "WARNING=!"
set "INSTALL=+"

REM Main execution
call :print_header
call :print_info "Starting GCM installation process..."
echo.

call :check_existing_installation
if !errorlevel! neq 0 exit /b !errorlevel!

call :detect_platform
if !errorlevel! neq 0 exit /b !errorlevel!

call :get_latest_version
if !errorlevel! neq 0 exit /b !errorlevel!

set "INSTALL_DIR=%USERPROFILE%\.local\bin"
call :print_info "Installation directory: %INSTALL_DIR%"
echo.

call :show_system_info
call :download_binary
if !errorlevel! neq 0 exit /b !errorlevel!

call :add_to_path
if !errorlevel! neq 0 exit /b !errorlevel!

call :verify_installation
call :show_completion

goto :eof

REM Functions start here

:show_help
echo GCM installer - GitHub Config Manager Installation Script for Windows
echo.
echo Usage: %~nx0 [OPTIONS]
echo.
echo Options:
echo   --quiet, -q         Run in quiet mode (minimal output)
echo   --version, -v VER   Install specific version (e.g., v1.0.0)
echo   --help, -h          Show this help message
echo.
echo Examples:
echo   %~nx0                  # Install latest version
echo   %~nx0 --quiet          # Install quietly
echo   %~nx0 --version v1.0.0 # Install specific version
goto :eof

:print_header
if %QUIET_MODE%==1 goto :eof
cls
call :print_separator "="
echo.
echo.
echo      ██████╗  ██████╗███╗   ███╗
echo     ██╔════╝ ██╔════╝████╗ ████║
echo     ██║  ███╗██║     ██╔████╔██║
echo     ██║   ██║██║     ██║╚██╔╝██║
echo     ╚██████╔╝╚██████╗██║ ╚═╝ ██║
echo      ╚═════╝  ╚═════╝╚═╝     ╚═╝
echo.
echo.
echo %BOLD%%WHITE%                GitHub Config Manager Installer%RESET%
echo %DIM%%GRAY%            Fast and secure installation process%RESET%
echo.
call :print_separator "="
echo.
goto :eof

:print_separator
set "char=%~1"
if "%char%"=="" set "char=-"
setlocal
set "line="
for /l %%i in (1,1,80) do set "line=!line!%char%"
echo !line!
endlocal
goto :eof

:print_info
if %QUIET_MODE%==1 goto :eof
echo %BLUE%%BOLD% %INFO%  INFO%RESET% %GRAY%^|%RESET% %~1
goto :eof

:print_success
if %QUIET_MODE%==1 goto :eof
echo %GREEN%%BOLD% %CHECKMARK%  SUCCESS%RESET% %GRAY%^|%RESET% %~1
goto :eof

:print_warning
echo %YELLOW%%BOLD% %WARNING%  WARNING%RESET% %GRAY%^|%RESET% %~1
goto :eof

:print_error
echo %RED%%BOLD% %CROSSMARK%  ERROR%RESET% %GRAY%^|%RESET% %~1
goto :eof

:print_step
if %QUIET_MODE%==1 goto :eof
echo %PURPLE%%BOLD% %ARROW%  STEP%RESET% %GRAY%^|%RESET% %~1
goto :eof

:detect_platform
call :print_step "Detecting system platform..."
set "PLATFORM_OS=windows"
set "PLATFORM_ARCH=amd64"

if "%PROCESSOR_ARCHITECTURE%"=="ARM64" set "PLATFORM_ARCH=arm64"
if "%PROCESSOR_ARCHITEW6432%"=="AMD64" set "PLATFORM_ARCH=amd64"

set "PLATFORM=%PLATFORM_OS%/%PLATFORM_ARCH%"
call :print_success "Detected platform: %PLATFORM%"
echo.
goto :eof

:get_latest_version
call :print_step "Fetching latest version information..."

if not "%SPECIFIC_VERSION%"=="" (
    set "LATEST_VERSION=%SPECIFIC_VERSION%"
    call :print_success "Using specified version: %LATEST_VERSION%"
    echo.
    goto :eof
)

REM Try to get latest version using PowerShell
for /f "tokens=*" %%i in ('powershell -NoProfile -Command "(Invoke-RestMethod -Uri 'https://api.github.com/repos/sijunda/github-config-manager/releases/latest' -TimeoutSec 30).tag_name" 2^>nul') do set "LATEST_VERSION=%%i"

if "%LATEST_VERSION%"=="" (
    call :print_error "Failed to get latest version information"
    call :print_info "Please check your internet connection or specify a version with --version"
    exit /b 1
)

call :print_success "Latest version: %LATEST_VERSION%"
echo.
goto :eof

:check_existing_installation
call :print_step "Checking for existing installation..."

set "EXISTING_FOUND=0"

REM Check if gcm is in PATH
where gcm >nul 2>&1
if %errorlevel%==0 (
    set "EXISTING_FOUND=1"
)

REM Check common install locations
if exist "%USERPROFILE%\.local\bin\gcm.exe" set "EXISTING_FOUND=1"

if %EXISTING_FOUND%==1 (
    echo.
    call :print_separator "-"
    echo %BOLD%%WHITE%Existing Installation Detected:%RESET%
    call :print_separator "-"
    
    where gcm >nul 2>&1
    if !errorlevel!==0 (
        for /f "tokens=*" %%v in ('gcm version 2^>nul') do (
            echo %GREEN% %CHECKMARK%%RESET% Command available: %BOLD%gcm%RESET% %DIM%^(%%v^)%RESET%
            goto :show_existing_done
        )
    )
    :show_existing_done
    
    if exist "%USERPROFILE%\.gcm" (
        echo %BLUE% %INFO%%RESET% Data directory: %BOLD%%USERPROFILE%\.gcm%RESET%
    )
    
    call :print_separator "-"
    echo.
    call :print_warning "GCM is already installed on this system!"
    echo.
    call :print_separator "-"
    echo %BOLD%%WHITE%What you can do:%RESET%
    echo  * Run 'gcm version' to check current version
    echo  * Run 'gcm --help' to see available commands
    echo  * Use the uninstaller script first if you need to reinstall
    echo  * Run 'gcm doctor' to check system health
    call :print_separator "-"
    echo.
    call :print_separator "="
    echo %DIM%%GRAY%Installation cancelled - gcm already exists%RESET%
    call :print_separator "="
    echo.
    exit /b 0
) else (
    call :print_success "No existing installation found - proceeding with fresh install"
    echo.
)
goto :eof

:show_system_info
call :print_separator "-"
echo %BOLD%%WHITE%System Information:%RESET%
call :print_separator "-"
echo %GREEN% %CHECKMARK%%RESET% Operating System: %BOLD%Windows%RESET%
echo %GREEN% %CHECKMARK%%RESET% Architecture: %BOLD%%PLATFORM_ARCH%%RESET%
echo %GREEN% %CHECKMARK%%RESET% Version: %BOLD%%LATEST_VERSION%%RESET%
echo %BLUE% %INFO%%RESET% Install Directory: %BOLD%%INSTALL_DIR%%RESET%
call :print_separator "-"
echo.
goto :eof

:download_binary
call :print_step "Downloading gcm %LATEST_VERSION% for %PLATFORM%..."

set "DOWNLOAD_URL=https://github.com/sijunda/github-config-manager/releases/download/%LATEST_VERSION%/gcm-%PLATFORM_OS%-%PLATFORM_ARCH%.exe"
set "BINARY_PATH=%INSTALL_DIR%\gcm.exe"

call :print_info "Download URL: %DOWNLOAD_URL%"

REM Create install directory
if not exist "%INSTALL_DIR%" mkdir "%INSTALL_DIR%"

REM Download using PowerShell
call :print_info "Downloading binary..."
powershell -NoProfile -Command "Invoke-WebRequest -Uri '%DOWNLOAD_URL%' -OutFile '%BINARY_PATH%' -TimeoutSec 60" 2>nul

if not exist "%BINARY_PATH%" (
    call :print_error "Failed to download gcm binary"
    exit /b 1
)

REM Verify file size
for %%A in ("%BINARY_PATH%") do set "FILE_SIZE=%%~zA"
if %FILE_SIZE% LSS 1048576 (
    call :print_warning "Binary file seems unusually small (%FILE_SIZE% bytes)"
)

REM Quick validation
"%BINARY_PATH%" version >nul 2>&1
if %errorlevel% neq 0 (
    call :print_error "Downloaded binary appears to be corrupted or invalid"
    del "%BINARY_PATH%" 2>nul
    exit /b 1
)

call :print_success "Downloaded gcm binary to %BINARY_PATH%"
echo.
goto :eof

:add_to_path
call :print_step "Configuring Windows environment..."

REM Check if already in PATH
echo %PATH% | find /i "%INSTALL_DIR%" >nul
if %errorlevel%==0 (
    call :print_info "Install directory already in PATH"
    goto :eof
)

REM Add to user PATH using PowerShell
powershell -NoProfile -Command "$userPath = [Environment]::GetEnvironmentVariable('PATH', 'User'); if ($userPath) { [Environment]::SetEnvironmentVariable('PATH', \"$userPath;%INSTALL_DIR%\", 'User') } else { [Environment]::SetEnvironmentVariable('PATH', '%INSTALL_DIR%', 'User') }"

if %errorlevel%==0 (
    call :print_success "Added %INSTALL_DIR% to user PATH"
    set "PATH=%PATH%;%INSTALL_DIR%"
) else (
    call :print_warning "Could not add to PATH automatically"
    call :print_info "Please add %INSTALL_DIR% to your PATH manually"
)
goto :eof

:verify_installation
call :print_step "Verifying installation..."

"%INSTALL_DIR%\gcm.exe" version >nul 2>&1
if %errorlevel%==0 (
    for /f "tokens=*" %%v in ('"%INSTALL_DIR%\gcm.exe" version 2^>nul') do (
        call :print_success "Installation verified: %%v"
        goto :eof
    )
) else (
    call :print_warning "Installation completed, but verification failed"
    echo.
    call :print_separator "-"
    echo %BOLD%%WHITE%Manual Steps Required:%RESET%
    echo  1. Restart your terminal
    echo  2. Try running 'gcm version'
    echo  3. If issues persist, run 'gcm init' manually
    call :print_separator "-"
    echo.
)
goto :eof

:show_completion
echo.
call :print_separator "="
echo.
echo %GREEN%%BOLD%  INSTALLATION SUCCESSFUL!%RESET%
echo.
call :print_separator "-"
echo %BOLD%%WHITE%What was installed:%RESET%
echo  * gcm binary
echo  * Windows PATH configuration
call :print_separator "-"
echo %BOLD%%WHITE%Next Steps:%RESET%
echo  1. Restart your terminal/Command Prompt
echo  2. Verify with 'gcm version'
echo  3. Initialize with 'gcm init'
echo  4. Create your first profile with 'gcm profile create ^<name^>'
call :print_separator "-"
echo %BOLD%%WHITE%Quick Commands:%RESET%
echo  * gcm profile create work   - Create a profile
echo  * gcm use work              - Switch to a profile
echo  * gcm ssh generate work     - Generate SSH key
echo  * gcm github login-oauth    - Authenticate with GitHub
echo  * gcm doctor                - Check system health
call :print_separator "-"
echo Welcome to GCM!
call :print_separator "="
echo.
goto :eof
