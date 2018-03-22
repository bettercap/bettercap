package modules

import (
	"fmt"
	"io"
	"time"

	"github.com/bettercap/bettercap/core"
	"github.com/bettercap/bettercap/log"
	"github.com/bettercap/bettercap/session"

	"github.com/adrianmo/go-nmea"
	"github.com/tarm/serial"
)

type GPS struct {
	session.SessionModule

	serialPort string
	baudRate   int
	serial     *serial.Port
}

func NewGPS(s *session.Session) *GPS {
	gps := &GPS{
		SessionModule: session.NewSessionModule("http.server", s),
		serialPort:    "/dev/ttyUSB0",
		baudRate:      19200,
	}

	gps.AddParam(session.NewStringParameter("gps.device",
		gps.serialPort,
		"",
		"Serial device of the GPS hardware."))

	gps.AddParam(session.NewIntParameter("gps.baudrate",
		fmt.Sprintf("%d", gps.baudRate),
		"Baud rate of the GPS serial device."))

	gps.AddHandler(session.NewModuleHandler("gps on", "",
		"Start acquiring from the GPS hardware.",
		func(args []string) error {
			return gps.Start()
		}))

	gps.AddHandler(session.NewModuleHandler("gps off", "",
		"Stop acquiring from the GPS hardware.",
		func(args []string) error {
			return gps.Stop()
		}))

	gps.AddHandler(session.NewModuleHandler("gps.show", "",
		"Show the last coordinates returned by the GPS hardware.",
		func(args []string) error {
			return gps.Show()
		}))

	return gps
}

func (gps *GPS) Name() string {
	return "gps"
}

func (gps *GPS) Description() string {
	return "A module talking with GPS hardware on a serial interface."
}

func (gps *GPS) Author() string {
	return "Simone Margaritelli <evilsocket@protonmail.com>"
}

func (gps *GPS) Configure() (err error) {
	if gps.Running() {
		return session.ErrAlreadyStarted
	} else if err, gps.serialPort = gps.StringParam("gps.device"); err != nil {
		return err
	} else if err, gps.baudRate = gps.IntParam("gps.baudrate"); err != nil {
		return err
	}

	gps.serial, err = serial.OpenPort(&serial.Config{
		Name:        gps.serialPort,
		Baud:        gps.baudRate,
		ReadTimeout: time.Second * 1,
	})

	return
}

func (gps *GPS) readLine() (line string, err error) {
	var n int

	b := make([]byte, 1)
	for {
		if n, err = gps.serial.Read(b); err != nil {
			return
		} else if n == 1 {
			if b[0] == '\n' {
				return core.Trim(line), nil
			} else {
				line += string(b[0])
			}
		}
	}
}

func (gps *GPS) Show() error {
	fmt.Printf("latitude:%f longitude:%f quality:%s satellites:%d altitude:%f\n",
		gps.Session.GPS.Latitude,
		gps.Session.GPS.Longitude,
		gps.Session.GPS.FixQuality,
		gps.Session.GPS.NumSatellites,
		gps.Session.GPS.Altitude)

	gps.Session.Refresh()

	return nil
}

func (gps *GPS) Start() error {
	if err := gps.Configure(); err != nil {
		return err
	}

	return gps.SetRunning(true, func() {

		defer gps.serial.Close()

		for gps.Running() {
			if line, err := gps.readLine(); err == nil {
				if info, err := nmea.Parse(line); err == nil {
					s := info.Sentence()
					// http://aprs.gids.nl/nmea/#gga
					if s.Type == "GNGGA" {
						gps.Session.GPS = info.(nmea.GNGGA)
					} else {
						log.Debug("Skipping message %s: %v", s.Type, s)
					}
				} else {
					log.Debug("Error parsing line '%s': %s", line, err)
				}
			} else if err != io.EOF {
				log.Warning("Error while reading serial port: %s", err)
			}
		}
	})
}

func (gps *GPS) Stop() error {
	return gps.SetRunning(false, func() {
		// let the read fail and exit
		gps.serial.Close()
	})
}
