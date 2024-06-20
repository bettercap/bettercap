@ECHO OFF
set TARGET_OS=windows
set TARGET_ARCH=amd64
set OUTPUT=bettercap.exe

rem CGO_CFLAGS="-I/Path-to-winpcap-x64-include-dir -I/Path-to-libusb-1.0-x64-include-dir"
set CGO_CFLAGS="-I/c/src/vcpkg/packages/winpcap_x64-windows-static/include -I/c/src/vcpkg/packages/libusb_x64-windows-static/include/libusb-1.0"

rem CGO_LDFLAGS="-L/Path-to-winpcap-x64-lib-dir -L/Path-to-libusb-1.0-x64-lib-dir"
set CGO_LDFLAGS="-L/c/src/vcpkg/packages/winpcap_x64-windows-static/lib -L/c/src/vcpkg/packages/libusb_x64-windows-static/lib"

rem Get deps
go get ./...

rem Build
go build -o bettercap.exe .