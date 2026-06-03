@echo off
REM  ── Lenovo Power Monitor Driver Build Script ──────────────────────────
REM  Requires: VS2022 + WDK 10.0.26100.0 (or compatible)
REM  Usage:   build.bat [x64|x86] [testsign]
REM  Example: build.bat x64 testsign

set TARGET=%1
if "%TARGET%"=="" set TARGET=x64

set SIGN=%2

REM ── Paths (auto-detect) ───────────────────────────────────────────────
set "VS=C:\Program Files\Microsoft Visual Studio\2022\Professional"
if not exist "%VS%" set "VS=C:\Program Files\Microsoft Visual Studio\2022\Community"
if not exist "%VS%" (
    echo [ERROR] Visual Studio 2022 not found.
    exit /b 1
)

set "MSVC=%VS%\VC\Tools\MSVC"
for /f "delims=" %%i in ('dir /b /ad /o-n "%MSVC%" 2^>nul') do (
    set "MSVC_VER=%%i"
    goto :found_msvc
)
echo [ERROR] MSVC directory not found under %MSVC%
exit /b 1
:found_msvc

set "MSVC_BIN=%MSVC%\%MSVC_VER%\bin\Hostx64\%TARGET%"
if not exist "%MSVC_BIN%" (
    echo [ERROR] MSVC bin path not found: %MSVC_BIN%
    exit /b 1
)

set "WDK=C:\Program Files (x86)\Windows Kits\10"
set "WDK_INC=%WDK%\Include\10.0.26100.0"
set "WDK_LIB=%WDK%\Lib\10.0.26100.0"
set "WDK_BIN=%WDK%\bin\10.0.26100.0\x64"
if /i not "%TARGET%"=="x64" set "WDK_BIN=%WDK%\bin\10.0.26100.0\%TARGET%"

echo [BUILD] Target: %TARGET%
echo [BUILD] MSVC:  %MSVC_BIN%
echo [BUILD] WDK:    %WDK_BIN%

REM ── Set env ─────────────────────────────────────────────────────────────
set "PATH=%MSVC_BIN%;%WDK_BIN%;%PATH%"
set "INCLUDE=%WDK_INC%\km\include;%WDK_INC%\shared;%WDK_INC%\um;%WDK_INC%\winrt"
set "LIB=%WDK_LIB%\km\%TARGET%;%WDK_LIB%\um\%TARGET%"

REM ── Compile ─────────────────────────────────────────────────────────────
echo [BUILD] Compiling driver.c ...
cd /d "%~dp0"

cl.exe /nologo /c ^
    /I"%WDK_INC%\km\include" ^
    /I"%WDK_INC%\shared" ^
    driver.c ^
    /Fo:driver.obj ^
    /W3 /WX- /O2 /Ob2 /Oy- ^
    /D "NDIS_WDM=1" /D "_WIN64" /D "WIN64" /D "WINVER=0x0A00" /D "_WIN32_WINNT=0x0A00" ^
    /kernel /TC /GS- /hotpatch

if errorlevel 1 (
    echo [ERROR] Compilation failed.
    exit /b 1
)
echo [BUILD] driver.obj OK

REM ── Link ───────────────────────────────────────────────────────────────
echo [BUILD] Linking lenovo_power.sys ...
link.exe /nologo ^
    /DRIVER /DRIVER:WDM ^
    /SUBSYSTEM:NATIVE,10.0 /VERSION:1.0 ^
    /MERGE:.rdata=.text ^
    /IGNORE:4198,4078 ^
    /ENTRY:GsDriverEntry ^
    /STACK:0x40000,0x1000 ^
    /PDB:lenovo_power.pdb ^
    /OUT:lenovo_power.sys ^
    /SUBsystem:NATIVE ^
    driver.obj ^
    "%WDK_LIB%\km\%TARGET%\ntoskrnl.lib" ^
    "%WDK_LIB%\km\%TARGET%\hal.lib" ^
    "%WDK_LIB%\km\%TARGET%\wmilib.lib"

if errorlevel 1 (
    echo [ERROR] Linking failed.
    exit /b 1
)
echo [BUILD] lenovo_power.sys OK

REM ── Test-sign (optional) ──────────────────────────────────────────────
if /i "%SIGN%"=="testsign" (
    echo [BUILD] Test-signing lenovo_power.sys ...
    "%WDK_BIN%\signtool.exe" sign ^
        /v /n "Lenovo" ^
        /tr http://timestamp.digicert.com ^
        /td sha256 ^
        /fd sha256 ^
        /a ^
        lenovo_power.sys
    if errorlevel 1 (
        echo [WARN]  signtool failed - try: bcdedit /set testsigning on
    ) else (
        echo [BUILD] Signing OK
    )
)

echo.
echo [BUILD] Done. Output: %CD%\lenovo_power.sys
exit /b 0
