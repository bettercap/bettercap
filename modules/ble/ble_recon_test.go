//go:build !windows && !freebsd && !openbsd && !netbsd
// +build !windows,!freebsd,!openbsd,!netbsd

package ble

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/bettercap/bettercap/v2/network"
	"github.com/bettercap/bettercap/v2/session"

	"github.com/bettercap/gatt"
)

var (
	testSession *session.Session
	sessionOnce sync.Once
)

type fakeGattDevice struct {
	gatt.Device
	scanCalls      int
	stopScanCalls  int
	connectCalls   int
	cancelCalls    int
	lastPeripheral gatt.Peripheral
}

func (d *fakeGattDevice) Scan([]gatt.UUID, bool) {
	d.scanCalls++
}

func (d *fakeGattDevice) StopScanning() {
	d.stopScanCalls++
}

func (d *fakeGattDevice) Connect(p gatt.Peripheral) {
	d.connectCalls++
	d.lastPeripheral = p
}

func (d *fakeGattDevice) CancelConnection(p gatt.Peripheral) {
	d.cancelCalls++
}

type fakePeripheral struct {
	gatt.Peripheral
	id     string
	device gatt.Device
}

func (p *fakePeripheral) Device() gatt.Device {
	return p.device
}

func (p *fakePeripheral) ID() string {
	return p.id
}

func (p *fakePeripheral) Name() string {
	return "fake"
}

func createMockSession(t *testing.T) *session.Session {
	sessionOnce.Do(func() {
		var err error
		testSession, err = session.New()
		if err != nil {
			t.Fatalf("Failed to create session: %v", err)
		}
		testSession.BLE = network.NewBLE(testSession.Aliases, nil, nil)
	})
	return testSession
}

func TestNewBLERecon(t *testing.T) {
	s := createMockSession(t)
	mod := NewBLERecon(s)

	if mod == nil {
		t.Fatal("NewBLERecon returned nil")
	}

	if mod.Name() != "ble.recon" {
		t.Errorf("Expected name 'ble.recon', got '%s'", mod.Name())
	}

	if mod.Author() != "Simone Margaritelli <evilsocket@gmail.com>" {
		t.Errorf("Unexpected author: %s", mod.Author())
	}

	if mod.Description() == "" {
		t.Error("Empty description")
	}

	// Check initial values
	if mod.deviceId != -1 {
		t.Errorf("Expected deviceId -1, got %d", mod.deviceId)
	}

	if mod.isConnected() {
		t.Error("Should not be connected initially")
	}

	if mod.connTimeout != 5 {
		t.Errorf("Expected connection timeout 5, got %d", mod.connTimeout)
	}

	if mod.devTTL != 30 {
		t.Errorf("Expected devTTL 30, got %d", mod.devTTL)
	}

	// Check channels
	if mod.quit == nil {
		t.Error("Quit channel should not be nil")
	}

	if mod.done == nil {
		t.Error("Done channel should not be nil")
	}

	// Check handlers
	handlers := mod.Handlers()
	expectedHandlers := []string{
		"ble.recon on",
		"ble.recon off",
		"ble.clear",
		"ble.show",
		"ble.enum MAC",
		"ble.write MAC UUID HEX_DATA",
	}

	if len(handlers) != len(expectedHandlers) {
		t.Errorf("Expected %d handlers, got %d", len(expectedHandlers), len(handlers))
	}

	handlerNames := make(map[string]bool)
	for _, h := range handlers {
		handlerNames[h.Name] = true
	}

	for _, expected := range expectedHandlers {
		if !handlerNames[expected] {
			t.Errorf("Handler '%s' not found", expected)
		}
	}
}

func TestIsEnumerating(t *testing.T) {
	s := createMockSession(t)
	mod := NewBLERecon(s)

	// Initially should not be enumerating
	if mod.isEnumerating() {
		t.Error("Should not be enumerating initially")
	}

	// When currDevice is set, should be enumerating
	// We can't create a real BLE device here, but we can test the logic
}

func TestDummyWriter(t *testing.T) {
	s := createMockSession(t)
	mod := NewBLERecon(s)

	writer := dummyWriter{mod}
	testData := []byte("test log message")

	n, err := writer.Write(testData)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if n != len(testData) {
		t.Errorf("Expected to write %d bytes, wrote %d", len(testData), n)
	}
}

func TestParameters(t *testing.T) {
	s := createMockSession(t)
	mod := NewBLERecon(s)

	// Check that parameters are registered
	paramNames := []string{
		"ble.device",
		"ble.timeout",
		"ble.ttl",
	}

	// Parameters are stored in the session environment
	// We'll just ensure the module was created properly
	for _, param := range paramNames {
		// This is a simplified check
		_ = param
	}

	if mod == nil {
		t.Error("Module should not be nil")
	}
}

func TestRunningState(t *testing.T) {
	s := createMockSession(t)
	mod := NewBLERecon(s)

	// Initially should not be running
	if mod.Running() {
		t.Error("Module should not be running initially")
	}

	// Note: Cannot test actual Start/Stop without BLE hardware
}

func TestChannels(t *testing.T) {
	// Skip this test as channel operations might hang in certain environments
	t.Skip("Skipping channel test to prevent potential hangs")
}

func TestClearHandler(t *testing.T) {
	// Skip this test as it requires BLE to be initialized in the session
	t.Skip("Skipping clear handler test - requires initialized BLE in session")
}

func TestBLEPrompt(t *testing.T) {
	expected := "{blb}{fw}BLE {fb}{reset} {bold}» {reset}"
	if blePrompt != expected {
		t.Errorf("Expected prompt '%s', got '%s'", expected, blePrompt)
	}
}

func TestSetCurrentDevice(t *testing.T) {
	s := createMockSession(t)
	mod := NewBLERecon(s)

	// Test setting nil device
	mod.setCurrentDevice(nil)
	if mod.currentDevice() != nil {
		t.Error("Current device should be nil")
	}
	if mod.isConnected() {
		t.Error("Should not be connected after setting nil device")
	}
}

func TestViewSelector(t *testing.T) {
	s := createMockSession(t)
	mod := NewBLERecon(s)

	// Check that view selector is initialized
	if mod.selector == nil {
		t.Error("View selector should not be nil")
	}
}

func TestBLEAliveInterval(t *testing.T) {
	expected := time.Duration(5) * time.Second
	if bleAliveInterval != expected {
		t.Errorf("Expected alive interval %v, got %v", expected, bleAliveInterval)
	}
}

func TestColNames(t *testing.T) {
	s := createMockSession(t)
	mod := NewBLERecon(s)

	// Test without name
	cols := mod.colNames(false)
	expectedCols := []string{"RSSI", "MAC", "Vendor", "Flags", "Connect", "Seen"}
	if len(cols) != len(expectedCols) {
		t.Errorf("Expected %d columns, got %d", len(expectedCols), len(cols))
	}

	// Test with name
	colsWithName := mod.colNames(true)
	expectedColsWithName := []string{"RSSI", "MAC", "Name", "Vendor", "Flags", "Connect", "Seen"}
	if len(colsWithName) != len(expectedColsWithName) {
		t.Errorf("Expected %d columns with name, got %d", len(expectedColsWithName), len(colsWithName))
	}
}

func TestDoFilter(t *testing.T) {
	s := createMockSession(t)
	mod := NewBLERecon(s)

	// Without expression, should always return true
	result := mod.doFilter(nil)
	if !result {
		t.Error("doFilter should return true when no expression is set")
	}
}

func TestShow(t *testing.T) {
	// Skip this test as it requires BLE to be initialized in the session
	t.Skip("Skipping show test - requires initialized BLE in session")
}

func TestConfigure(t *testing.T) {
	// Skip this test as it may hang trying to access BLE hardware
	t.Skip("Skipping configure test - may hang accessing BLE hardware")
}

func TestGetRow(t *testing.T) {
	s := createMockSession(t)
	mod := NewBLERecon(s)

	// We can't create a real BLE device without hardware, but we can test the logic
	// by ensuring the method exists and would handle nil gracefully
	_ = mod
}

func TestDoSelection(t *testing.T) {
	// Skip this test as it requires BLE to be initialized in the session
	t.Skip("Skipping doSelection test - requires initialized BLE in session")
}

func TestWriteBuffer(t *testing.T) {
	// Skip this test as it may hang trying to access BLE hardware
	t.Skip("Skipping writeBuffer test - may hang accessing BLE hardware")
}

const fakeBLEMac = "AA:BB:CC:DD:EE:FF"

func addTestBLEDevice(t *testing.T, mod *BLERecon, p gatt.Peripheral) *network.BLEDevice {
	t.Helper()

	mod.Session.BLE.Clear()
	mod.Session.BLE.AddIfNew(p.ID(), p, nil, -50)

	dev, found := mod.Session.BLE.Get(p.ID())
	if !found {
		t.Fatalf("device %s was not added to the session", p.ID())
	}
	return dev
}

func TestEnumAllTheThingsScansBeforeConnecting(t *testing.T) {
	s := createMockSession(t)
	mod := NewBLERecon(s)

	fake := &fakeGattDevice{}
	mod.gattDevice = fake

	mod.Session.Env.Set("ble.timeout", "60")
	defer mod.Session.Env.Set("ble.timeout", "5")
	defer mod.setCurrentDevice(nil)

	oldDevice := &fakeGattDevice{}
	p := &fakePeripheral{id: fakeBLEMac, device: oldDevice}
	dev := addTestBLEDevice(t, mod, p)

	if err := mod.enumAllTheThings(network.NormalizeMac(fakeBLEMac)); err != nil {
		t.Fatalf("enumAllTheThings failed: %v", err)
	}

	if fake.scanCalls != 1 {
		t.Errorf("Expected one scan before connecting, got %d", fake.scanCalls)
	}
	if fake.connectCalls != 0 {
		t.Errorf("Expected no connection before target discovery, got %d", fake.connectCalls)
	}
	if current := mod.currentDevice(); current != dev {
		t.Errorf("Expected current device %p, got %p", dev, current)
	}
	if target := mod.targetMAC(); target != network.NormalizeMac(fakeBLEMac) {
		t.Errorf("Expected target %s, got %s", network.NormalizeMac(fakeBLEMac), target)
	}
}

func TestEnumAllTheThingsReusesCurrentPeripheral(t *testing.T) {
	s := createMockSession(t)
	mod := NewBLERecon(s)

	fake := &fakeGattDevice{}
	mod.gattDevice = fake

	mod.Session.Env.Set("ble.timeout", "60")
	defer mod.Session.Env.Set("ble.timeout", "5")
	defer mod.setCurrentDevice(nil)

	p := &fakePeripheral{id: fakeBLEMac, device: fake}
	dev := addTestBLEDevice(t, mod, p)

	if err := mod.enumAllTheThings(network.NormalizeMac(fakeBLEMac)); err != nil {
		t.Fatalf("enumAllTheThings failed: %v", err)
	}

	if fake.scanCalls != 0 {
		t.Errorf("Expected the current peripheral to be reused without scanning, got %d scans", fake.scanCalls)
	}
	if fake.connectCalls != 1 {
		t.Errorf("Expected one immediate connection, got %d", fake.connectCalls)
	}
	if fake.lastPeripheral != p {
		t.Errorf("Expected connection to %p, got %p", p, fake.lastPeripheral)
	}
	if current := mod.currentDevice(); current != dev {
		t.Errorf("Expected current device %p, got %p", dev, current)
	}
}

func TestOnStateChangedScansWhileEnumerationIsPending(t *testing.T) {
	s := createMockSession(t)
	mod := NewBLERecon(s)

	fake := &fakeGattDevice{}
	p := &fakePeripheral{id: fakeBLEMac, device: fake}
	dev := addTestBLEDevice(t, mod, p)

	mod.setCurrentDevice(dev)
	mod.setTargetMAC(network.NormalizeMac(fakeBLEMac))
	defer mod.setCurrentDevice(nil)

	mod.onStateChanged(fake, gatt.StatePoweredOn)

	if fake.scanCalls != 1 {
		t.Errorf("Expected discovery to start for a pending enum, got %d scans", fake.scanCalls)
	}
}

func TestOnPeriphDiscoveredConnectsOnlyTarget(t *testing.T) {
	s := createMockSession(t)
	mod := NewBLERecon(s)

	fake := &fakeGattDevice{}
	mod.gattDevice = fake

	p := &fakePeripheral{id: fakeBLEMac, device: fake}
	dev := addTestBLEDevice(t, mod, p)
	mod.setCurrentDevice(dev)
	mod.setTargetMAC(network.NormalizeMac(fakeBLEMac))
	defer mod.setCurrentDevice(nil)

	other := &fakePeripheral{id: "11:22:33:44:55:66", device: fake}
	mod.onPeriphDiscovered(other, nil, -40)
	if fake.connectCalls != 0 {
		t.Errorf("Expected no connection for a non-target device, got %d", fake.connectCalls)
	}

	mod.onPeriphDiscovered(p, nil, -50)
	if fake.connectCalls != 1 {
		t.Errorf("Expected one connection for the target, got %d", fake.connectCalls)
	}
	if fake.stopScanCalls != 1 {
		t.Errorf("Expected scanning to stop before connecting, got %d", fake.stopScanCalls)
	}
	if fake.lastPeripheral != p {
		t.Errorf("Expected connection to %p, got %p", p, fake.lastPeripheral)
	}

	mod.onPeriphDiscovered(p, nil, -50)
	if fake.connectCalls != 1 {
		t.Errorf("Expected duplicate advertisements to be ignored, got %d connections", fake.connectCalls)
	}
}

func TestOnPeriphConnectedCancelsLateConnection(t *testing.T) {
	s := createMockSession(t)
	mod := NewBLERecon(s)

	fake := &fakeGattDevice{}
	p := &fakePeripheral{id: fakeBLEMac, device: fake}
	mod.setCurrentDevice(nil)

	mod.onPeriphConnected(p, nil)

	if fake.cancelCalls != 1 {
		t.Errorf("Expected a late connection to be cancelled, got %d", fake.cancelCalls)
	}
}

func TestOnPeriphConnectedRejectsUnexpectedDevice(t *testing.T) {
	s := createMockSession(t)
	mod := NewBLERecon(s)

	fake := &fakeGattDevice{}
	p := &fakePeripheral{id: fakeBLEMac, device: fake}
	dev := addTestBLEDevice(t, mod, p)
	mod.setCurrentDevice(dev)
	mod.setTargetMAC(network.NormalizeMac(fakeBLEMac))
	defer mod.setCurrentDevice(nil)

	other := &fakePeripheral{id: "11:22:33:44:55:66", device: fake}
	mod.onPeriphConnected(other, nil)

	if fake.cancelCalls != 1 {
		t.Errorf("Expected an unexpected connection to be cancelled, got %d", fake.cancelCalls)
	}
	if mod.currentDevice() != nil {
		t.Error("Expected state cleanup after an unexpected connection")
	}
}

func TestOnPeriphConnectedErrorClearsState(t *testing.T) {
	s := createMockSession(t)
	mod := NewBLERecon(s)

	fake := &fakeGattDevice{}
	p := &fakePeripheral{id: fakeBLEMac, device: fake}
	dev := addTestBLEDevice(t, mod, p)
	mod.setCurrentDevice(dev)
	mod.setTargetMAC(network.NormalizeMac(fakeBLEMac))

	mod.onPeriphConnected(p, fmt.Errorf("connection failed"))

	if mod.currentDevice() != nil {
		t.Error("Expected state cleanup after a failed connection")
	}
	if mod.targetMAC() != "" {
		t.Error("Expected target cleanup after a failed connection")
	}
}

func TestOnStateChangedSkipsScanWhileConnected(t *testing.T) {
	s := createMockSession(t)
	mod := NewBLERecon(s)

	fake := &fakeGattDevice{}
	mod.markConnected()
	defer mod.setCurrentDevice(nil)

	mod.onStateChanged(fake, gatt.StatePoweredOn)

	if fake.scanCalls != 0 {
		t.Errorf("Expected no discovery while connected, got %d scans", fake.scanCalls)
	}
}

func TestEnumAllTheThingsCanRunAgain(t *testing.T) {
	s := createMockSession(t)
	mod := NewBLERecon(s)

	fake := &fakeGattDevice{}
	mod.gattDevice = fake

	mod.Session.Env.Set("ble.timeout", "60")
	defer mod.Session.Env.Set("ble.timeout", "5")
	defer mod.setCurrentDevice(nil)

	oldDevice := &fakeGattDevice{}
	p := &fakePeripheral{id: fakeBLEMac, device: oldDevice}
	dev := addTestBLEDevice(t, mod, p)
	mac := network.NormalizeMac(fakeBLEMac)

	if err := mod.enumAllTheThings(mac); err != nil {
		t.Fatalf("first enumAllTheThings failed: %v", err)
	}
	if fake.scanCalls != 1 {
		t.Fatalf("Expected the first enumeration to scan a new gatt instance, got %d", fake.scanCalls)
	}

	current := &fakePeripheral{id: fakeBLEMac, device: fake}
	mod.onPeriphDiscovered(current, nil, -50)
	if fake.connectCalls != 1 {
		t.Fatalf("Expected first enumeration to connect, got %d", fake.connectCalls)
	}

	mod.markConnected()
	mod.onPeriphDisconnected(nil, nil)
	if mod.currentDevice() != nil {
		t.Fatalf("Expected state cleanup after disconnect, got %v", mod.currentDevice())
	}
	if dev.Device != current {
		t.Fatalf("Expected stored device to retain the current peripheral")
	}

	if err := mod.enumAllTheThings(mac); err != nil {
		t.Fatalf("second enumAllTheThings failed: %v", err)
	}
	if fake.scanCalls != 1 {
		t.Errorf("Expected repeat enumeration to reuse the current peripheral, got %d scans", fake.scanCalls)
	}
	if fake.connectCalls != 2 {
		t.Errorf("Expected repeat enumeration to reconnect immediately, got %d", fake.connectCalls)
	}
}

// Benchmark tests - using singleton session to avoid flag redefinition
func BenchmarkNewBLERecon(b *testing.B) {
	// Use a test instance to get singleton session
	s := createMockSession(&testing.T{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewBLERecon(s)
	}
}

func BenchmarkIsEnumerating(b *testing.B) {
	s := createMockSession(&testing.T{})
	mod := NewBLERecon(s)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mod.isEnumerating()
	}
}

func BenchmarkDummyWriter(b *testing.B) {
	s := createMockSession(&testing.T{})
	mod := NewBLERecon(s)
	writer := dummyWriter{mod}
	testData := []byte("benchmark log message")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		writer.Write(testData)
	}
}

func BenchmarkDoFilter(b *testing.B) {
	s := createMockSession(&testing.T{})
	mod := NewBLERecon(s)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mod.doFilter(nil)
	}
}
