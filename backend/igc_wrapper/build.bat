@echo off
REM Build igc_wrapper.dll using MSVC x64
REM Requires Visual Studio 2019/2022 with C++ workload

REM Initialize MSVC environment
set "VCVARS="
for %%p in (
    "C:\Program Files\Microsoft Visual Studio\2022\Community\VC\Auxiliary\Build\vcvarsall.bat"
    "C:\Program Files\Microsoft Visual Studio\2022\Professional\VC\Auxiliary\Build\vcvarsall.bat"
    "C:\Program Files\Microsoft Visual Studio\2022\Enterprise\VC\Auxiliary\Build\vcvarsall.bat"
    "C:\Program Files (x86)\Microsoft Visual Studio\2019\Community\VC\Auxiliary\Build\vcvarsall.bat"
    "C:\Program Files (x86)\Microsoft Visual Studio\2019\Professional\VC\Auxiliary\Build\vcvarsall.bat"
) do (
    if exist %%p (
        set "VCVARS=%%p"
        goto :found
    )
)
echo ERROR: Visual Studio not found
exit /b 1

:found
call %VCVARS% x64

cl.exe /nologo /O2 /W3 /EHsc /LD ^
    igc_wrapper.cpp ^
    /Fe:igc_wrapper.dll ^
    /link /DEF:igc_wrapper.def /IMPLIB:igc_wrapper.lib

if %ERRORLEVEL% EQU 0 (
    echo SUCCESS: igc_wrapper.dll built
) else (
    echo FAILED: build error %ERRORLEVEL%
    exit /b %ERRORLEVEL%
)
