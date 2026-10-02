@echo off
setlocal EnableExtensions
title GoWA - PM2 stop

set "PM2_APP=API-whatsapp"

echo Deteniendo GoWA en PM2 (%PM2_APP%)...

where pm2 >nul 2>&1
if errorlevel 1 (
    echo ERROR: PM2 no esta en PATH. npm install -g pm2
    pause
    exit /b 1
)

call pm2 describe %PM2_APP% >nul 2>&1
if errorlevel 1 (
    echo No hay proceso registrado como "%PM2_APP%".
    call pm2 status
    pause
    exit /b 0
)

call pm2 stop %PM2_APP%
call pm2 delete %PM2_APP%

echo.
echo GoWA detenido. Para iniciar: PM2-start.bat
call pm2 status
pause
endlocal
