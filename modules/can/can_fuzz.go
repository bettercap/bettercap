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

func (mod *CANModule) fuzzSelectFrame(id string, rng *rand.Rand) (uint64, error) {
	// let's try as an hex number first
	frameID, err := strconv.ParseUint(id, 16, 32)
	if err != nil {
		// not a number, use as node name if we have a dbc
		if mod.dbc.Loaded() {
			fromSender := mod.dbc.MessagesBySender(id)
			if len(fromSender) == 0 {
				return 0, fmt.Errorf("no messages defined in DBC file for node %s, available nodes: %s", id, mod.dbc.Senders())
			}

			idx := rng.Intn(len(fromSender))
			selected := fromSender[idx]
			mod.Info("selected %s > (%d) %s", id, selected.ID, selected.Name)
			frameID = uint64(selected.ID)
		} else {
			// no dbc, just return the error
			return 0, err
		}
	}
	return frameID, nil
}

func (mod *CANModule) fuzzGenerateFrame(frameID uint64, size int, rng *rand.Rand) (*can.Frame, error) {
	dataLen := 0
	frameData := ([]byte)(nil)

	// if we have a DBC
	if mod.dbc.Loaded() {
		if message := mod.dbc.MessageById(uint32(frameID)); message != nil {
			mod.Info("using message %s", message.Name)
			dataLen = int(message.Length)
			frameData = make([]byte, dataLen)
			if _, err := rand.Read(frameData); err != nil {
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("message with id %d not found in DBC file, available ids: %v", frameID, strings.Join(mod.dbc.AvailableMessages(), ", "))
		}
	} else {
		if size <= 0 {
			// pick randomly
			dataLen = rng.Intn(int(can.MaxDataLength))
		} else {
			// user selected
			dataLen = size
		}
		frameData = make([]byte, dataLen)
		if _, err := rand.Read(frameData); err != nil {
			return nil, err
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

	return &frame, nil
}

func (mod *CANModule) Fuzz(id string, optSize string) error {
	rncSource := rand.NewSource(time.Now().Unix())
	rng := rand.New(rncSource)

	fuzzSize := 0
	if optSize != "" {
		if num, err := strconv.Atoi(optSize); err != nil {
			return fmt.Errorf("could not parse numeric size from '%s': %v", optSize, err)
		} else if num > 8 {
			return fmt.Errorf("max can frame size is 8, %d given", num)
		} else {
			fuzzSize = num
		}
	}

	if frameID, err := mod.fuzzSelectFrame(id, rng); err != nil {
		return err
	} else if frame, err := mod.fuzzGenerateFrame(frameID, fuzzSize, rng); err != nil {
		return err
	} else {
		mod.Info("injecting %s of CAN frame %d ...",
			humanize.Bytes(uint64(frame.Length)), frame.ID)
		if err := mod.send.TransmitFrame(context.Background(), *frame); err != nil {
			return err
		}
	}
	return nil
}
