@echo off
setlocal EnableExtensions
title GoWA - REST + MCP

set "ROOT=%~dp0"
if "%ROOT:~-1%"=="\" set "ROOT=%ROOT:~0,-1%"
set "SRC=%ROOT%\src"

if not exist "%SRC%\go.mod" (
    echo No se encontro el modulo Go en: %SRC%
    pause
    exit /b 1
)

cd /d "%SRC%"

if not exist ".env" (
    echo [AVISO] No hay .env en src\. Copia .env.example si aun no lo hiciste.
    echo.
)

set "GOWA_PORT=3000"
for /f "usebackq tokens=1,* delims==" %%a in (`findstr /b /i "APP_PORT=" ".env" 2^>nul`) do (
    set "GOWA_PORT=%%b"
)

echo Iniciando Go WhatsApp Web Multi-Device ^(rest^)...
echo Directorio: %CD%
echo UI/API: http://localhost:%GOWA_PORT%
echo Docs:     http://localhost:%GOWA_PORT%/docs/
echo.

where go >nul 2>&1
if not errorlevel 1 (
    echo Compilando con purego ^(SQLite sin CGO^)...
    go build -tags purego -o go-whatsapp-web-multidevice.exe .
    if errorlevel 1 (
        echo Fallo go build. Probando go run...
        go run -tags purego . rest
        goto :done
    )
)

if exist "go-whatsapp-web-multidevice.exe" (
    echo Usando: go-whatsapp-web-multidevice.exe rest
    go-whatsapp-web-multidevice.exe rest
    goto :done
)

if exist "whatsapp.exe" (
    echo Usando: whatsapp.exe rest
    whatsapp.exe rest
    goto :done
)

echo No hay Go en PATH ni binario compilado.
pause
exit /b 1

:done
echo.
if errorlevel 1 (
    echo El servidor termino con error ^(codigo %errorlevel%^).
) else (
    echo El servidor se detuvo.
)
pause
endlocal
