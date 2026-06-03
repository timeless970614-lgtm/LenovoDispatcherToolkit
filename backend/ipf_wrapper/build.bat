@echo off
call "C:\Program Files\Microsoft Visual Studio\2022\Community\VC\Auxiliary\Build\vcvars64.bat"
cd /d "C:\LenovoDispatcher\LenovoToolkit\backend\ipf_wrapper"
cl /LD /EHsc /D_CRT_SECURE_NO_WARNINGS ipf_wrapper.cpp ipf_wrapper.def /Fe:ipf_wrapper.dll
if %ERRORLEVEL%==0 (
    echo SUCCESS: ipf_wrapper.dll built
) else (
    echo FAILED: error code %ERRORLEVEL%
)
