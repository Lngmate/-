@echo off
title LabShare Frontend Server [Port 5173]
echo ========================================================
echo   Lab Equipment Reservation System - Frontend Server
echo   Port: 5173  URL: http://localhost:5173
echo ========================================================
cd /d "%~dp0frontend"
where npm.cmd >nul 2>&1
if errorlevel 1 (
    echo [ERROR] Node.js/npm was not found. Install Node.js and reopen this launcher.
    pause
    exit /b 1
)
if not exist "node_modules\.bin\vite.cmd" (
    echo [INFO] Installing frontend dependencies ...
    call npm.cmd ci
    if errorlevel 1 (
        echo [ERROR] Dependency installation failed. Check the npm output above.
        pause
        exit /b 1
    )
)
call npm.cmd run dev
pause
