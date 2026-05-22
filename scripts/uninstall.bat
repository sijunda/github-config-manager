@echo off
setlocal enabledelayedexpansion

REM GCM (GitHub Config Manager) uninstallation script for Windows Command Prompt
REM This script removes gcm from %USERPROFILE%\.local\bin and cleans PATH

REM Parse command line arguments
set SHOW_HELP=0

:parse_args
if "%~1"=="" goto :args_done
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
set "TRASH=DEL"
set "WARNING=!"
set "QUESTION=?"
set "STOP=STOP"
set "CLEAN=CLN"
set "SHIELD=KEEP"
set "INFO=i"

REM Main execution
call :print_header
call :print_info "Starting GCM uninstallation process..."
echo.

call :check_gcm_installation
set "INSTALLATION_FOUND=!errorlevel!"

if !INSTALLATION_FOUND!==0 (
    call :print_warning "GCM does not appear to be installed on this system"
    echo.
    call :print_separator "-"
    echo %BOLD%%WHITE%No GCM installation found!%RESET%
    call :print_separator "-"
    echo It looks like GCM is not installed or has already been removed.
    echo Common reasons:
    echo  * GCM was never installed
    echo  * GCM was already uninstalled
    echo  * GCM was installed in a different location
    call :print_separator "-"
    echo.
    set /p "RESPONSE=Do you want to clean any remaining traces? (y/N): "
    if /i "!RESPONSE!" neq "y" (
        echo.
        call :print_info "Exiting without making changes"
        call :print_separator "="
        echo %DIM%%GRAY%No changes were made to your system.%RESET%
        call :print_separator "="
        echo.
        exit /b 0
    )
    echo.
    call :print_info "Proceeding with cleanup of any remaining traces..."
    echo.
) else (
    call :print_success "GCM installation detected"
    echo.
)

call :show_uninstall_options

set /p "RESPONSE=Choose an option (1/2/3/4): "
echo.

if "!RESPONSE!"=="1" (
    call :minimal_removal
) else if "!RESPONSE!"=="2" (
    call :complete_removal
) else if "!RESPONSE!"=="3" (
    call :nuclear_clean
) else (
    echo.
    call :print_info "Uninstallation cancelled by user"
    call :print_separator "="
    echo %DIM%%GRAY%No changes were made to your system.%RESET%
    call :print_separator "="
    echo.
)

goto :eof

REM Functions start here

:show_help
echo GCM uninstaller - GitHub Config Manager Uninstallation Script for Windows
echo.
echo Usage: %~nx0 [OPTIONS]
echo.
echo Options:
echo   --help, -h          Show this help message
echo.
echo Examples:
echo   %~nx0               # Run interactive uninstaller
echo   %~nx0 --help        # Show help
goto :eof

:print_header
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
echo %BOLD%%WHITE%              GitHub Config Manager Uninstaller%RESET%
echo %DIM%%GRAY%            Safe and complete uninstallation process%RESET%
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
echo %BLUE%%BOLD% %INFO%  INFO%RESET% %GRAY%^|%RESET% %~1
goto :eof

:print_success
echo %GREEN%%BOLD% %CHECKMARK%  SUCCESS%RESET% %GRAY%^|%RESET% %~1
goto :eof

:print_warning
echo %YELLOW%%BOLD% %WARNING%  WARNING%RESET% %GRAY%^|%RESET% %~1
goto :eof

:print_error
echo %RED%%BOLD% %CROSSMARK%  ERROR%RESET% %GRAY%^|%RESET% %~1
goto :eof

:print_step
echo %PURPLE%%BOLD% %ARROW%  STEP%RESET% %GRAY%^|%RESET% %~1
goto :eof

:check_gcm_installation
call :print_step "Checking GCM installation..."

set "INSTALL_DIR=%USERPROFILE%\.local\bin"
set "GOPATH_BIN=%GOPATH%\bin"
if "%GOPATH_BIN%"=="\bin" set "GOPATH_BIN=%USERPROFILE%\go\bin"
set "GCM_DIR=%USERPROFILE%\.gcm"
set "FOUND=0"

REM Check binary in multiple locations
if exist "%INSTALL_DIR%\gcm.exe" (
    set "FOUND=1"
    echo %GREEN% %CHECKMARK%%RESET% Binary found: %BOLD%%INSTALL_DIR%\gcm.exe%RESET%
)
if exist "%GOPATH_BIN%\gcm.exe" (
    set "FOUND=1"
    echo %GREEN% %CHECKMARK%%RESET% Binary found: %BOLD%%GOPATH_BIN%\gcm.exe%RESET%
)

REM Check if gcm is in PATH
where gcm >nul 2>&1
if !errorlevel!==0 (
    set "FOUND=1"
    for /f "tokens=*" %%v in ('gcm version 2^>nul') do (
        echo %GREEN% %CHECKMARK%%RESET% Command available: %BOLD%gcm%RESET% %DIM%^(%%v^)%RESET%
    )
)

REM Check data directory
if exist "%GCM_DIR%" (
    set "FOUND=1"
    echo %BLUE% %INFO%%RESET% Data directory: %BOLD%%GCM_DIR%%RESET%
)

if !FOUND!==0 (
    exit /b 0
) else (
    exit /b 1
)

:show_uninstall_options
call :print_separator "="
echo %BOLD%%WHITE% %QUESTION%  UNINSTALLATION OPTIONS%RESET%
call :print_separator "="
echo.
echo %CYAN%%BOLD%1)%RESET% %WHITE%Minimal Removal%RESET% %DIM%(Recommended)%RESET%
echo    * Remove gcm binary
echo    * Clean PATH configuration
echo    * %GREEN%Keep%RESET% profiles, tokens, SSH keys, and configuration
echo.
echo %RED%%BOLD%2)%RESET% %WHITE%Complete Removal%RESET% %DIM%(Permanent)%RESET%
echo    * Remove gcm binary
echo    * Clean PATH configuration
echo    * %RED%Delete%RESET% all profiles and configuration
echo    * %RED%Delete%RESET% encrypted tokens, backup archives, audit logs
echo.
echo %RED%%BOLD%3)%RESET% %WHITE%Nuclear Clean%RESET% %DIM%(Everything - no trace left)%RESET%
echo    * Everything in option 2, plus:
echo    * %RED%Delete%RESET% git global identity (user.name, user.email, signingkey)
echo    * %RED%Delete%RESET% git credential config for github.com
echo    * %RED%Delete%RESET% git local identity and GCM markers in current repo
echo.
echo %GRAY%%BOLD%4)%RESET% %WHITE%Cancel%RESET%
echo    * Exit without making any changes
echo.
call :print_separator "-"
goto :eof

:minimal_removal
call :print_info "Proceeding with minimal removal..."
echo.

call :print_separator "-"
echo %YELLOW%%BOLD% %STOP%  FINAL CONFIRMATION%RESET%
call :print_separator "-"
set /p "CONFIRM=Proceed with minimal removal? (y/N): "

if /i "!CONFIRM!" neq "y" (
    echo.
    call :print_info "Uninstallation cancelled by user"
    call :print_separator "="
    echo %DIM%%GRAY%No changes were made to your system.%RESET%
    call :print_separator "="
    echo.
    goto :eof
)

echo.
call :remove_binary
echo.
call :remove_from_path
echo.
call :show_completion_minimal
goto :eof

:complete_removal
call :print_info "Proceeding with complete removal..."
echo.

call :print_separator "-"
echo %RED%%BOLD% %STOP%  DANGER: COMPLETE REMOVAL%RESET%
call :print_separator "-"
echo %RED%This will permanently delete ALL GCM data including profiles, tokens, and backups!%RESET%
call :print_separator "-"
set /p "CONFIRM=Type 'DELETE' to confirm complete removal: "

if "!CONFIRM!" neq "DELETE" (
    echo.
    call :print_info "Uninstallation cancelled - confirmation text did not match"
    call :print_separator "="
    echo %DIM%%GRAY%No changes were made to your system.%RESET%
    call :print_separator "="
    echo.
    goto :eof
)

echo.
call :remove_binary
echo.
call :remove_from_path
echo.
call :remove_git_credential
echo.
call :remove_gcm_dir
echo.
call :show_completion_complete
goto :eof

:remove_binary
set "INSTALL_DIR=%USERPROFILE%\.local\bin"
set "GOPATH_BIN=%GOPATH%\bin"
if "%GOPATH_BIN%"=="\bin" set "GOPATH_BIN=%USERPROFILE%\go\bin"
call :print_step "Removing gcm binary..."

set "BINARY_REMOVED=0"

if exist "%INSTALL_DIR%\gcm.exe" (
    del "%INSTALL_DIR%\gcm.exe" 2>nul
    if !errorlevel!==0 (
        call :print_success "Removed gcm from %INSTALL_DIR%"
        set "BINARY_REMOVED=1"
    ) else (
        call :print_error "Failed to remove binary from %INSTALL_DIR%"
    )
)

if exist "%GOPATH_BIN%\gcm.exe" (
    del "%GOPATH_BIN%\gcm.exe" 2>nul
    if !errorlevel!==0 (
        call :print_success "Removed gcm from %GOPATH_BIN%"
        set "BINARY_REMOVED=1"
    ) else (
        call :print_error "Failed to remove binary from %GOPATH_BIN%"
    )
)

if exist "C:\usr\local\bin\gcm.exe" (
    del "C:\usr\local\bin\gcm.exe" 2>nul
    if !errorlevel!==0 (
        call :print_success "Removed gcm from C:\usr\local\bin"
        set "BINARY_REMOVED=1"
    ) else (
        call :print_error "Failed to remove binary from C:\usr\local\bin"
    )
)

if !BINARY_REMOVED!==0 (
    call :print_warning "gcm binary not found in expected locations"
)
goto :eof

:remove_from_path
set "INSTALL_DIR=%USERPROFILE%\.local\bin"
call :print_step "Cleaning PATH configuration..."

REM Remove from user PATH using PowerShell
powershell -NoProfile -Command "$userPath = [Environment]::GetEnvironmentVariable('PATH', 'User'); if ($userPath -like '*%INSTALL_DIR%*') { $parts = $userPath -split ';' | Where-Object { $_ -ne '%INSTALL_DIR%' -and $_ -ne '' }; $newPath = $parts -join ';'; [Environment]::SetEnvironmentVariable('PATH', $newPath, 'User'); Write-Output 'removed' } else { Write-Output 'notfound' }" > "%TEMP%\gcm_path_result.txt" 2>nul

set /p "PATH_RESULT=" < "%TEMP%\gcm_path_result.txt"
del "%TEMP%\gcm_path_result.txt" 2>nul

if "!PATH_RESULT!"=="removed" (
    call :print_success "Removed %INSTALL_DIR% from user PATH"
) else (
    call :print_info "No GCM PATH configuration found"
)
goto :eof

:remove_gcm_dir
set "GCM_DIR=%USERPROFILE%\.gcm"
call :print_step "Removing GCM data directory..."

if exist "%GCM_DIR%" (
    rmdir /s /q "%GCM_DIR%" 2>nul
    if !errorlevel!==0 (
        call :print_success "Removed GCM data directory"
    ) else (
        call :print_error "Failed to remove data directory"
    )
) else (
    call :print_warning "GCM directory not found at %GCM_DIR%"
)
goto :eof

:remove_git_credential
call :print_step "Cleaning git credential config..."

REM Check and remove credential helper
for /f "tokens=*" %%a in ('git config --global "credential.https://github.com.helper" 2^>nul') do (
    git config --global --unset-all "credential.https://github.com.helper" 2>nul
    call :print_success "Removed credential.https://github.com.helper"
)

REM Check and remove credential username
for /f "tokens=*" %%a in ('git config --global "credential.https://github.com.username" 2^>nul') do (
    git config --global --unset-all "credential.https://github.com.username" 2>nul
    call :print_success "Removed credential.https://github.com.username"
)
goto :eof

:show_completion_minimal
echo.
call :print_separator "="
echo.
echo %GREEN%%BOLD%  MINIMAL UNINSTALLATION COMPLETE!%RESET%
echo.
call :print_separator "-"
echo %BOLD%%WHITE%What was removed:%RESET%
echo  * gcm binary
echo  * PATH configuration
echo.
echo %BOLD%%WHITE%What was kept:%RESET%
echo  * Profiles and configuration in ~\.gcm
echo  * SSH keys (in ~\.ssh)
echo  * Encrypted tokens and backup archives
call :print_separator "-"
echo %BOLD%%WHITE%Final Steps:%RESET%
echo  1. Restart your terminal/Command Prompt
echo  2. Verify with 'gcm version' (should show error)
echo  3. Manually remove '%%USERPROFILE%%\.gcm' if you change your mind later
call :print_separator "-"
echo Thank you for using GCM!
call :print_separator "="
echo.
goto :eof

:show_completion_complete
echo.
call :print_separator "="
echo.
echo %GREEN%%BOLD%  COMPLETE UNINSTALLATION SUCCESSFUL!%RESET%
echo.
call :print_separator "-"
echo %BOLD%%WHITE%What was removed:%RESET%
echo  * gcm binary
echo  * PATH configuration
echo  * All profiles and configuration
echo  * Encrypted tokens and audit logs
echo  * Backup archives
call :print_separator "-"
echo %BOLD%%WHITE%Final Steps:%RESET%
echo  1. Restart your terminal/Command Prompt
echo  2. Verify with 'gcm version' (should show error)
call :print_separator "-"
echo Thank you for using GCM!
call :print_separator "="
echo.
goto :eof

:nuclear_clean
call :print_info "Proceeding with NUCLEAR clean..."
echo.

call :print_separator "-"
echo %RED%%BOLD% %STOP%  DANGER: NUCLEAR CLEAN - NO TRACE LEFT%RESET%
call :print_separator "-"
echo %RED%This will permanently delete EVERYTHING: binary, data, git identity, credentials!%RESET%
call :print_separator "-"
set /p "CONFIRM=Type 'NUKE' to confirm nuclear clean: "

if "!CONFIRM!" neq "NUKE" (
    echo.
    call :print_info "Uninstallation cancelled - confirmation text did not match"
    call :print_separator "="
    echo %DIM%%GRAY%No changes were made to your system.%RESET%
    call :print_separator "="
    echo.
    goto :eof
)

echo.
call :remove_binary
echo.
call :remove_from_path
echo.
call :remove_git_identity
echo.
call :remove_git_credential
echo.
call :remove_gcm_dir
echo.
call :show_completion_nuclear
goto :eof

:remove_git_identity
call :print_step "Removing git identity configuration..."

for /f "tokens=*" %%a in ('git config --global user.name 2^>nul') do (
    git config --global --unset-all user.name 2>nul
    call :print_success "Unset git global user.name"
)
for /f "tokens=*" %%a in ('git config --global user.email 2^>nul') do (
    git config --global --unset-all user.email 2>nul
    call :print_success "Unset git global user.email"
)
for /f "tokens=*" %%a in ('git config --global user.signingkey 2^>nul') do (
    git config --global --unset-all user.signingkey 2>nul
    call :print_success "Unset git global user.signingkey"
)
for /f "tokens=*" %%a in ('git config --global commit.gpgsign 2^>nul') do (
    git config --global --unset-all commit.gpgsign 2>nul
    call :print_success "Unset git global commit.gpgsign"
)

REM Clean local repo if inside one
git rev-parse --is-inside-work-tree >nul 2>&1
if !errorlevel!==0 (
    git config --local --unset-all user.name 2>nul
    git config --local --unset-all user.email 2>nul
    git config --local --unset-all user.signingkey 2>nul
    git config --local --unset-all commit.gpgsign 2>nul
    call :print_success "Cleaned git local identity"
    for /f "tokens=*" %%r in ('git rev-parse --show-toplevel 2^>nul') do (
        if exist "%%r\.gcm-profile" del "%%r\.gcm-profile" 2>nul
        if exist "%%r\.git\gcm-session" del "%%r\.git\gcm-session" 2>nul
    )
    call :print_success "Removed GCM markers"
)
goto :eof

:show_completion_nuclear
echo.
call :print_separator "="
echo.
echo %GREEN%%BOLD%  NUCLEAR CLEAN SUCCESSFUL - NO TRACE LEFT!%RESET%
echo.
call :print_separator "-"
echo %BOLD%%WHITE%What was removed:%RESET%
echo  * gcm binary (from all locations)
echo  * PATH configuration
echo  * Git global identity (user.name, user.email, signingkey, gpgsign)
echo  * Git local identity and GCM markers
echo  * Git credential config for github.com
echo  * All profiles, tokens, config, backups, cache
call :print_separator "-"
echo %BOLD%%WHITE%Final Steps:%RESET%
echo  1. Restart your terminal/Command Prompt
echo  2. Verify with 'where gcm' (should show error)
call :print_separator "-"
echo Thank you for using GCM!
call :print_separator "="
echo.
goto :eof
