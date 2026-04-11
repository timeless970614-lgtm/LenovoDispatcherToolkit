@echo off
cd /d "%~dp0"
call "C:\Program Files\Microsoft Visual Studio\18\Community\VC\Auxiliary\Build\vcvarsall.bat" x64
cl /EHsc /W4 /O2 test.cpp /Fe:test3.exe user32.lib
pause
