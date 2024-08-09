package can

import (
	"context"

	"github.com/dustin/go-humanize"
	"go.einride.tech/can"
)

func (mod *CANModule) Inject(expr string) (err error) {
	frame := can.Frame{}
	if err := frame.UnmarshalString(expr); err != nil {
		return err
	}

	mod.Info("injecting %s of CAN frame %d ...",
		humanize.Bytes(uint64(frame.Length)), frame.ID)

	if err := mod.send.TransmitFrame(context.Background(), frame); err != nil {
		return err
	}

	return
}
