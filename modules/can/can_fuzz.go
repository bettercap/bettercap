package can

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"go.einride.tech/can"
)

func (mod *CANModule) Fuzz(id string) error {
	rncSource := rand.NewSource(time.Now().Unix())
	rng := rand.New(rncSource)

	// let's try as an hex number first
	// frameID, err := strconv.Atoi(id)
	frameID, err := strconv.ParseUint(id, 16, 32)
	dataLen := 0
	frameData := ([]byte)(nil)

	if err != nil {
		if mod.dbc != nil {
			// not a number, use as node name
			fromSender := mod.dbc.MessagesBySender(id)
			if len(fromSender) == 0 {
				return fmt.Errorf("no messages defined in DBC file for node %s, available nodes: %s", id, mod.dbc.Senders())
			}

			idx := rng.Intn(len(fromSender))
			selected := fromSender[idx]
			mod.Info("selected %s > (%d) %s", id, selected.ID, selected.Name)
			frameID = uint64(selected.ID)
		} else {
			return err
		}
	}

	// if we have a DBC
	if mod.dbc.Loaded() {
		if message := mod.dbc.MessageById(uint32(frameID)); message != nil {
			mod.Info("found as %s", message.Name)

			dataLen = int(message.Length)
			frameData = make([]byte, dataLen)
			if _, err := rand.Read(frameData); err != nil {
				return err
			}
		} else {
			avail := []string{}
			for _, msg := range mod.dbc.Messages() {
				avail = append(avail, fmt.Sprintf("%d (%s)", msg.ID, msg.Name))
			}
			return fmt.Errorf("message with id %d not found in DBC file, available ids: %v", frameID, strings.Join(avail, ", "))
		}
	} else {
		dataLen = rng.Intn(int(can.MaxDataLength))
		frameData = make([]byte, dataLen)

		if _, err := rand.Read(frameData); err != nil {
			return err
		}

		mod.Warning("no dbc loaded, creating frame with %d bytes of random data", dataLen)
	}

	frame := can.Frame{
		ID:         uint32(frameID),
		Length:     uint8(dataLen),
		IsRemote:   false,
		IsExtended: false,
	}

	copy(frame.Data[:], frameData)

	mod.Info("injecting %s of CAN frame %d ...",
		humanize.Bytes(uint64(frame.Length)), frame.ID)

	if err := mod.send.TransmitFrame(context.Background(), frame); err != nil {
		return err
	}

	return nil
}
