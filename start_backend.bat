@echo off
title LabShare Backend Server [Port 8080]
echo ========================================================
echo   Lab Equipment Reservation System - Backend Server
echo   Port: 8080 ^| DB: SQLite
echo ========================================================
cd /d "%~dp0backend"
if exist server.exe (
    echo [OK] Launching server.exe ...
    server.exe
) else (
    where go >nul 2>&1
    if not errorlevel 1 (
        echo [OK] Launching with installed Go ...
        go run ./cmd/server
    ) else if exist "..\go\bin\go.exe" (
        echo [OK] Launching with portable Go ...
        "..\go\bin\go.exe" run ./cmd/server
    ) else (
        echo [ERROR] Go was not found. Install Go and reopen this launcher.
        pause
        exit /b 1
    )
)
pause
