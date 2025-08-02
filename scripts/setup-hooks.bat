@echo off
REM Script to set up Git hooks for the project (Windows version)
REM Run this script once after cloning the repository

echo Setting up Git hooks...

REM Check if we're in a Git repository
if not exist ".git" (
    echo Error: Not a Git repository
    exit /b 1
)

REM Create .git/hooks directory if it doesn't exist
if not exist ".git\hooks" mkdir ".git\hooks"

REM Copy hooks
if exist ".githooks\pre-commit" (
    copy ".githooks\pre-commit" ".git\hooks\pre-commit" >nul
    echo Pre-commit hook installed
) else (
    echo Error: Pre-commit hook file not found
)

if exist ".githooks\pre-push" (
    copy ".githooks\pre-push" ".git\hooks\pre-push" >nul
    echo Pre-push hook installed
) else (
    echo Error: Pre-push hook file not found
)

REM Try to set up Git hooks path (Git 2.9+)
git config core.hooksPath .githooks >nul 2>&1
if %errorlevel% equ 0 (
    echo Git hooks path configured to use .githooks directory
) else (
    echo Warning: Could not set hooks path. Using manual copy method.
)

echo.
echo Git hooks setup complete!
echo.
echo The following hooks are now active:
echo   • pre-commit: Runs formatting and syntax checks on staged files
echo   • pre-push: Runs full tests and linting before pushing
echo.
echo To bypass hooks temporarily:
echo   • git commit --no-verify
echo   • git push --no-verify
echo.
echo Make sure to install golangci-lint for full linting support:
echo    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

pause