package hid_recon

const (
	frameDelay = 12
)

var (
	helloData     = []byte{0x00, 0x4F, 0x00, 0x04, 0xB0, 0x10, 0x00, 0x00, 0x00, 0xED}
	keepAliveData = []byte{0x00, 0x40, 0x04, 0xB0, 0x0C}
)

type LogitechBuilder struct {
}

func (b LogitechBuilder) frameFor(cmd *Command) []byte {
	data := []byte{0, 0xC1, cmd.Mode, cmd.HID, 0, 0, 0, 0, 0, 0}
	sz := len(data)
	last := sz - 1
	sum := byte(0xff)

	for i := 0; i < last; i++ {
		sum = (sum - data[i]) & 0xff
	}
	sum = (sum + 1) & 0xff
	data[last] = sum

	return data
}

func (b LogitechBuilder) BuildFrames(commands []*Command) error {
	numCommands := len(commands)
	for i, cmd := range commands {
		if i == 0 {
			cmd.AddFrame(helloData, frameDelay)
		}

		next := (*Command)(nil)
		if i < numCommands-1 {
			next = commands[i+1]
		}

		if cmd.IsHID() {
			cmd.AddFrame(b.frameFor(cmd), frameDelay)
			cmd.AddFrame(keepAliveData, 0)
			if next == nil || cmd.HID == next.HID || next.IsSleep() {
				cmd.AddFrame(b.frameFor(&Command{}), 0)
			}
		} else if cmd.IsSleep() {
			for i, num := 0, cmd.Sleep/10; i < num; i++ {
				cmd.AddFrame(keepAliveData, 10)
			}
		}
	}

	return nil
}
