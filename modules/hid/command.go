package hid

import (
	"time"
)

type Frame struct {
	Data  []byte
	Delay time.Duration
}

func NewFrame(buf []byte, delay int) Frame {
	return Frame{
		Data:  buf,
		Delay: time.Millisecond * time.Duration(delay),
	}
}

type Command struct {
	Mode   byte
	HID    byte
	Sleep  int
	Frames []Frame
}

func (cmd *Command) AddFrame(buf []byte, delay int) {
	if cmd.Frames == nil {
		cmd.Frames = make([]Frame, 0)
	}
	cmd.Frames = append(cmd.Frames, NewFrame(buf, delay))
}

func (cmd Command) IsHID() bool {
	return cmd.HID != 0 || cmd.Mode != 0
}

func (cmd Command) IsSleep() bool {
	return cmd.Sleep > 0
}
