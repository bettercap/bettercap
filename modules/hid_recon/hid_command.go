package hid_recon

type Frame []byte

type Command struct {
	Mode   byte
	HID    byte
	Char   string
	Sleep  byte
	Frames []Frame
}
