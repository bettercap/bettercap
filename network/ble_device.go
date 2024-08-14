//go:build !windows
// +build !windows

package network

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/evilsocket/islazy/str"
	"tinygo.org/x/bluetooth"
)

type BLECharacteristic struct {
	UUID       string      `json:"uuid"`
	Name       string      `json:"name"`
	MTU        uint16      `json:"mtu"`
	Properties []string    `json:"properties"`
	Data       interface{} `json:"data"`
}

func NewBLECharacteristic(UUID string) BLECharacteristic {
	char := BLECharacteristic{
		UUID:       UUID,
		Properties: make([]string, 0),
	}
	if name, found := BLE_Characteristics[char.UUID]; found {
		char.Name = name
	}
	return char
}

type BLEService struct {
	UUID            string              `json:"uuid"`
	Name            string              `json:"name"`
	Characteristics []BLECharacteristic `json:"characteristics"`
}

func NewBLEService(UUID string) BLEService {
	service := BLEService{
		UUID:            UUID,
		Characteristics: make([]BLECharacteristic, 0),
	}
	if name, found := BLE_Services[service.UUID]; found {
		service.Name = name
	}
	return service
}

type BLEDevice struct {
	sync.RWMutex

	Alias            string
	LastSeen         time.Time
	Address          string
	Name             string
	Vendor           string
	RSSI             int16
	Advertisement    []byte
	ManufacturerData []bluetooth.ManufacturerDataElement
	ServiceData      []bluetooth.ServiceDataElement
	Services         []BLEService
}

type bleDeviceJSON struct {
	LastSeen    time.Time    `json:"last_seen"`
	Name        string       `json:"name"`
	MAC         string       `json:"mac"`
	Alias       string       `json:"alias"`
	Vendor      string       `json:"vendor"`
	RSSI        int16        `json:"rssi"`
	Connectable bool         `json:"connectable"`
	Flags       string       `json:"flags"`
	Services    []BLEService `json:"services"`
}

func NewBLEDevice(scanResult bluetooth.ScanResult) *BLEDevice {
	devAddress := scanResult.Address.String()
	devName := scanResult.LocalName()
	devAdv := scanResult.Bytes()
	return &BLEDevice{
		LastSeen:         time.Now(),
		Address:          devAddress,
		Name:             devName,
		RSSI:             scanResult.RSSI,
		Advertisement:    devAdv,
		Vendor:           ManufLookup(devAddress),
		ManufacturerData: make([]bluetooth.ManufacturerDataElement, 0),
		ServiceData:      make([]bluetooth.ServiceDataElement, 0),
		Services:         make([]BLEService, 0),
	}
}

func sliceContains(slice []interface{}, elem interface{}) bool {
	for _, item := range slice {
		if item == elem {
			return true
		}
	}
	return false
}

func (dev *BLEDevice) hasManu(manu bluetooth.ManufacturerDataElement) bool {
	for _, item := range dev.ManufacturerData {
		if reflect.DeepEqual(item, manu) {
			return true
		}
	}
	return false
}

func (dev *BLEDevice) hasSvc(svc bluetooth.ServiceDataElement) bool {
	for _, item := range dev.ServiceData {
		if reflect.DeepEqual(item, svc) {
			return true
		}
	}
	return false
}

func (dev *BLEDevice) ResetServices() {
	dev.Lock()
	defer dev.Unlock()
	dev.Services = make([]BLEService, 0)
}

func (dev *BLEDevice) AddService(svc BLEService) {
	dev.Lock()
	defer dev.Unlock()
	dev.Services = append(dev.Services, svc)
}

func (dev *BLEDevice) Update(scanResult bluetooth.ScanResult, alias string) {
	dev.Lock()
	defer dev.Unlock()

	devName := str.Trim(scanResult.LocalName())
	devAdv := scanResult.Bytes()
	devManu := scanResult.ManufacturerData()
	devService := scanResult.ServiceData()

	dev.LastSeen = time.Now()
	dev.RSSI = scanResult.RSSI

	if alias != "" {
		dev.Alias = alias
	}

	if devName != "" {
		dev.Name = devName
	}

	if devAdv != nil && !reflect.DeepEqual(devAdv, dev.Advertisement) {
		dev.Advertisement = devAdv
	}

	for _, manu := range devManu {
		if !dev.hasManu(manu) {
			dev.ManufacturerData = append(dev.ManufacturerData, manu)
		}
	}

	for _, svc := range devService {
		if !dev.hasSvc(svc) {
			dev.ServiceData = append(dev.ServiceData, svc)
		}
	}

	if dev.Vendor == "" || dev.Vendor[0] == '<' {
		for _, manu := range dev.ManufacturerData {
			if company, found := BLE_Companies[manu.CompanyID]; found {
				dev.Vendor = company
				break
			} else {
				dev.Vendor = fmt.Sprintf("<0x%x>", manu.CompanyID)
			}
		}
	}
}

func (d *BLEDevice) MarshalJSON() ([]byte, error) {
	d.RLock()
	defer d.RUnlock()

	doc := bleDeviceJSON{
		LastSeen:    d.LastSeen,
		Name:        d.Name,
		MAC:         d.Address,
		Alias:       d.Alias,
		Vendor:      d.Vendor,
		RSSI:        d.RSSI,
		Connectable: true,
		Flags:       "",
		Services:    d.Services,
	}
	return json.Marshal(doc)
}
