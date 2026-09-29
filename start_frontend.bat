@echo off
title LabShare Frontend Server [Port 5173]
echo ========================================================
echo   Lab Equipment Reservation System - Frontend Server
echo   Port: 5173  URL: http://localhost:5173
echo ========================================================
cd /d "%~dp0"
cd frontend
npm run dev
pause