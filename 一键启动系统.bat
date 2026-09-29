@echo off
title LabShare All-in-One Launcher
echo ========================================================
echo   Lab Equipment Reservation System (LabShare)
echo   Starting Backend (8080) and Frontend (5173)...
echo ========================================================
echo [1/3] Starting Backend Server in new window...
start "LabShare Backend Server [8080]" cmd /k "%~dp0start_backend.bat"

echo [2/3] Waiting 2 seconds...
timeout /t 2 /nobreak > nul

echo [3/3] Starting Frontend Server in new window...
start "LabShare Frontend Server [5173]" cmd /k "%~dp0start_frontend.bat"

timeout /t 3 /nobreak > nul
echo Opening browser at http://localhost:5173/ ...
start http://localhost:5173/