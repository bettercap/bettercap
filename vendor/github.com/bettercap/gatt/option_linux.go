package gatt

import (
	"errors"
	"io"

	"github.com/bettercap/gatt/linux/cmd"
)

// LnxDeviceID specifies which HCI device to use.
// If n is set to -1, all the available HCI devices will be probed.
// If chk is set to true, LnxDeviceID checks the LE support in the feature list of the HCI device.
// This is to filter devices that does not support LE. In case some LE driver that doesn't correctly
// set the LE support in its feature list, user can turn off the check.
// This option can only be used with NewDevice on Linux implementation.
func LnxDeviceID(n int, chk bool) Option {
	return func(d Device) error {
		d.(*device).devID = n
		d.(*device).chkLE = chk
		return nil
	}
}

// LnxMaxConnections is an optional parameter.
// If set, it overrides the default max connections supported.
// This option can only be used with NewDevice on Linux implementation.
func LnxMaxConnections(n int) Option {
	return func(d Device) error {
		d.(*device).maxConn = n
		return nil
	}
}

// LnxSetAdvertisingEnable sets the advertising data to the HCI device.
// This option can be used with Option on Linux implementation.
func LnxSetAdvertisingEnable(en bool) Option {
	return func(d Device) error {
		dd := d.(*device)
		if dd == nil {
			return errors.New("device is not initialized")
		}
		if err := dd.update(); err != nil {
			return err
		}
		return dd.hci.SetAdvertiseEnable(en)
	}
}

// LnxSetAdvertisingData sets the advertising data to the HCI device.
// This option can be used with NewDevice or Option on Linux implementation.
func LnxSetAdvertisingData(c *cmd.LESetAdvertisingData) Option {
	return func(d Device) error {
		d.(*device).advData = c
		return nil
	}
}

// LnxSetScanResponseData sets the scan response data to the HXI device.
// This option can be used with NewDevice or Option on Linux implementation.
func LnxSetScanResponseData(c *cmd.LESetScanResponseData) Option {
	return func(d Device) error {
		d.(*device).scanResp = c
		return nil
	}
}

// LnxSetAdvertisingParameters sets the advertising parameters to the HCI device.
// This option can be used with NewDevice or Option on Linux implementation.
func LnxSetAdvertisingParameters(c *cmd.LESetAdvertisingParameters) Option {
	return func(d Device) error {
		d.(*device).advParam = c
		return nil
	}
}

// LnxSendHCIRawCommand sends a raw command to the HCI device
// This option can be used with NewDevice or Option on Linux implementation.
func LnxSendHCIRawCommand(c cmd.CmdParam, rsp io.Writer) Option {
	return func(d Device) error {
		b, err := d.(*device).SendHCIRawCommand(c)
		if rsp == nil {
			return err
		}
		rsp.Write(b)
		return err
	}
}
