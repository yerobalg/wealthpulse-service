@echo off
REM Windows cmd equivalent of the Makefile.
REM Usage: make.bat ^<target^>
REM   targets: install-goose, install-swag, migrate, swagger, run, build

setlocal EnableDelayedExpansion

if "%~1"=="" (
    set "TARGET=run"
) else (
    set "TARGET=%~1"
)

if /I "%TARGET%"=="install-goose" goto install_goose
if /I "%TARGET%"=="install-swag"  goto install_swag
if /I "%TARGET%"=="migrate"       goto migrate
if /I "%TARGET%"=="swagger"       goto swagger
if /I "%TARGET%"=="run"           goto run
if /I "%TARGET%"=="build"         goto build

echo Unknown target: %TARGET%
echo Available: install-goose, install-swag, migrate, swagger, run, build
exit /b 1

:install_goose
go install github.com/pressly/goose/v3/cmd/goose@latest
exit /b %ERRORLEVEL%

:install_swag
go install github.com/swaggo/swag/cmd/swag@latest
exit /b %ERRORLEVEL%

:migrate
call :load_env
goose -dir docs/sql postgres "host=%DB_HOST% port=%DB_PORT% user=%DB_USERNAME% password=%DB_PASSWORD% dbname=%DB_DBNAME% sslmode=disable" up
exit /b %ERRORLEVEL%

:swagger
swag init -o ./docs/api
exit /b %ERRORLEVEL%

:run
go run main.go
exit /b %ERRORLEVEL%

:build
go build -o bin/dantal-service.exe main.go
exit /b %ERRORLEVEL%

:load_env
if not exist ".env" exit /b 0
for /f "usebackq tokens=* eol=#" %%A in (".env") do (
    set "LINE=%%A"
    if not "!LINE!"=="" (
        for /f "tokens=1,* delims==" %%K in ("!LINE!") do (
            set "VAL=%%L"
            REM strip surrounding double quotes if present
            if defined VAL (
                if "!VAL:~0,1!"=="\"" set "VAL=!VAL:~1!"
                if defined VAL if "!VAL:~-1!"=="\"" set "VAL=!VAL:~0,-1!"
            )
            set "%%K=!VAL!"
        )
    )
)
exit /b 0
