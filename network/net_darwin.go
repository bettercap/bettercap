package network

/*
#cgo LDFLAGS: -framework CoreWLAN -framework Foundation
#include <stdbool.h>
#include <stdlib.h>

const char *GetSupportedFrequencies(const char *iface);
bool SetInterfaceChannel(const char *iface, int channel);
*/
import "C"

import (
	"encoding/json"
	"errors"
	"net"
	"unsafe"
)

func getInterfaceName(iface net.Interface) string {
	return iface.Name
}

func ForceMonitorMode(iface string) error {
	return nil
}

func SetInterfaceChannel(iface string, channel int) error {
	cIface := C.CString(iface)
	defer C.free(unsafe.Pointer(cIface))

	success := C.SetInterfaceChannel(cIface, C.int(channel))
	if !success {
		return errors.New("failed to set interface channel")
	}

	SetInterfaceCurrentChannel(iface, channel)
	return nil
}

func GetSupportedFrequencies(iface string) ([]int, error) {
	cIface := C.CString(iface)
	defer C.free(unsafe.Pointer(cIface))

	cFrequencies := C.GetSupportedFrequencies(cIface)
	if cFrequencies == nil {
		return nil, errors.New("failed to get supported frequencies")
	}
	defer C.free(unsafe.Pointer(cFrequencies))

	frequenciesStr := C.GoString(cFrequencies)
	var frequencies []int
	err := json.Unmarshal([]byte(frequenciesStr), &frequencies)
	if err != nil {
		return nil, err
	}

	return frequencies, nil
}
