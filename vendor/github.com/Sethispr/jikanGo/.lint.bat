@echo off

echo Formatting Code
go fmt ./...

echo.
echo Tidying Modules
go mod tidy

echo.
echo Running Linter
golangci-lint run ./...

if %ERRORLEVEL% EQU 0 (
    echo.
    echo [NICE] Checks passed
) else (
    echo.
    echo [ERR] Linter issues ^^
    pause
)

echo.
echo Press any key to exit..
pause >nul
