@echo off
call "C:\Program Files\Microsoft Visual Studio\2022\Community\VC\Auxiliary\Build\vcvars64.bat"

echo === Active Scheme ===
powercfg /getactivescheme

echo.
echo === All Schemes ===
powercfg /list

echo.
echo === Processor Subgroup GUID ===
echo GUID: 54533251-82be-4824-96c1-47b60b740d00

echo.
echo === EPP (P-Core) ===
powercfg /query SCHEME_CURRENT 54533251-82be-4824-96c1-47b60b740d00 36687f9e-e3a5-4dbf-b1dc-15eb381c6863

echo.
echo === EPP (E-Core) ===
powercfg /query SCHEME_CURRENT 54533251-82be-4824-96c1-47b60b740d00 36687f9e-e3a5-4dbf-b1dc-15eb381c6864

echo.
echo === Hetero Inc ===
powercfg /query SCHEME_CURRENT 54533251-82be-4824-96c1-47b60b740d00 06cadf0e-64ed-448a-8927-ce7bf90eb35d

echo.
echo === Hetero Dec ===
powercfg /query SCHEME_CURRENT 54533251-82be-4824-96c1-47b60b740d00 12a0ab44-fe28-4fa9-b3bd-4b64f44960a6

echo.
echo === Max Freq P-Core ===
powercfg /query SCHEME_CURRENT 54533251-82be-4824-96c1-47b60b740d00 75b0ae3f-bce0-45a7-8c89-c9611c25e100

echo.
echo === Max Freq E-Core ===
powercfg /query SCHEME_CURRENT 54533251-82be-4824-96c1-47b60b740d00 75b0ae3f-bce0-45a7-8c89-c9611c25e101

echo.
echo === Soft Park ===
powercfg /query SCHEME_CURRENT 54533251-82be-4824-96c1-47b60b740d00 97cfac41-2217-47eb-992d-618b1977c907

echo.
echo === CP Min Cores ===
powercfg /query SCHEME_CURRENT 54533251-82be-4824-96c1-47b60b740d00 893dee8e-2bef-41e0-89c6-b55d0929964c

echo.
echo === Max Perf ===
powercfg /query SCHEME_CURRENT 54533251-82be-4824-96c1-47b60b740d00 bc5038f7-23e0-4960-96da-33abaf5935ec

pause
