REM
REM Automatically generate pcap_version.h based on pcap_version.h.in
REM for Windows
REM The version string comes from VERSION
REM @echo off
REM

setlocal enableextensions disabledelayedexpansion

set "search=%%%%LIBPCAP_VERSION%%%%"
set /p replace=<%1

if exist %3 del %3 2>nul

for /f "delims=" %%i in ('type %2' ) do (
	set "line=%%i"
	setlocal enabledelayedexpansion
	set "line=!line:%search%=%replace%!"
	>>%3 echo(!line!
	endlocal
)

echo pcap_version.h generated
