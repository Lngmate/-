@echo off
title LabShare Backend Server [Port 8080]
echo ========================================================
echo   Lab Equipment Reservation System - Backend Server
echo   Port: 8080 ^| DB: SQLite
echo ========================================================
cd /d "%~dp0"
cd backend
if exist server.exe (
    echo [OK] Launching server.exe ...
    server.exe
) else (
    echo [OK] Launching with Go ...
    ..\go\bin\go.exe run cmd/server/main.go
)
pause