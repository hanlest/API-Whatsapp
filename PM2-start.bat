@echo off
setlocal EnableExtensions EnableDelayedExpansion
title GoWA - PM2 start

REM PM2-start-all.bat pasa "nopause" para no bloquear el arranque del resto.
if /I "%~1"=="nopause" set "NO_PAUSE=1"

set "ROOT=%~dp0"
if "%ROOT:~-1%"=="\" set "ROOT=%ROOT:~0,-1%"
set "SRC=%ROOT%\src"
set "PM2_APP=API-whatsapp"
set "EXE=%SRC%\go-whatsapp-web-multidevice.exe"

echo [%date% %time%] PM2-start.bat
echo Raiz: %ROOT%

if not exist "%SRC%\go.mod" (
    echo ERROR: No se encontro go.mod en %SRC%
    goto :fail
)

where pm2 >nul 2>&1
if errorlevel 1 (
    echo ERROR: PM2 no esta en PATH.
    echo Instalalo con: npm install -g pm2
    goto :fail
)

cd /d "%SRC%" || goto :fail
echo Trabajando en: %CD%

if not exist ".env" (
    echo [AVISO] No hay .env en src\ — copia .env.example si hace falta.
)

set "GOWA_PORT=3000"
if exist ".env" (
    for /f "usebackq tokens=1,* delims==" %%a in (`findstr /b /i "APP_PORT=" ".env" 2^>nul`) do set "GOWA_PORT=%%b"
)

where go >nul 2>&1
if not errorlevel 1 (
    echo Compilando con purego...
    call go build -tags purego -o go-whatsapp-web-multidevice.exe .
    if errorlevel 1 (
        echo ERROR: fallo go build.
        goto :fail
    )
) else (
    echo [AVISO] Go no esta en PATH; se usa el .exe existente si hay.
)

if not exist "%EXE%" (
    echo ERROR: No existe %EXE%
    echo Instala Go o compila manualmente desde src\
    goto :fail
)

call pm2 describe %PM2_APP% >nul 2>&1
if errorlevel 1 (
    echo Iniciando %PM2_APP% en PM2...
    call pm2 start "%EXE%" --name %PM2_APP% --cwd "%SRC%" -- rest
) else (
    echo Reiniciando %PM2_APP% en PM2...
    call pm2 restart %PM2_APP% --update-env
)
if errorlevel 1 (
    echo ERROR: PM2 no pudo iniciar la aplicacion.
    goto :fail
)

echo.
echo OK — GoWA en PM2: %PM2_APP%
echo UI/API: http://localhost:!GOWA_PORT!
echo Docs:   http://localhost:!GOWA_PORT!/docs/
echo.
call pm2 status
echo.
echo Logs: pm2 logs %PM2_APP%
goto :end

:fail
echo.
echo PM2-start finalizo con error.
if not defined NO_PAUSE pause
exit /b 1

:end
if not defined NO_PAUSE pause
endlocal
exit /b 0
