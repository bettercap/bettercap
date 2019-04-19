package gps

import (
	"fmt"
	"io"
	"time"

	"github.com/bettercap/bettercap/session"

	"github.com/adrianmo/go-nmea"
	"github.com/tarm/serial"

	"github.com/evilsocket/islazy/str"
)

type GPS struct {
	session.SessionModule

	serialPort string
	baudRate   int
	serial     *serial.Port
}

func NewGPS(s *session.Session) *GPS {
	mod := &GPS{
		SessionModule: session.NewSessionModule("gps", s),
		serialPort:    "/dev/ttyUSB0",
		baudRate:      4800,
	}

	mod.AddParam(session.NewStringParameter("gps.device",
		mod.serialPort,
		"",
		"Serial device of the GPS hardware."))

	mod.AddParam(session.NewIntParameter("gps.baudrate",
		fmt.Sprintf("%d", mod.baudRate),
		"Baud rate of the GPS serial device."))

	mod.AddHandler(session.NewModuleHandler("gps on", "",
		"Start acquiring from the GPS hardware.",
		func(args []string) error {
			return mod.Start()
		}))

	mod.AddHandler(session.NewModuleHandler("gps off", "",
		"Stop acquiring from the GPS hardware.",
		func(args []string) error {
			return mod.Stop()
		}))

	mod.AddHandler(session.NewModuleHandler("gps.show", "",
		"Show the last coordinates returned by the GPS hardware.",
		func(args []string) error {
			return mod.Show()
		}))

	return mod
}

func (mod *GPS) Name() string {
	return "gps"
}

func (mod *GPS) Description() string {
	return "A module talking with GPS hardware on a serial interface."
}

func (mod *GPS) Author() string {
	return "Simone Margaritelli <evilsocket@gmail.com>"
}

func (mod *GPS) Configure() (err error) {
	if mod.Running() {
		return session.ErrAlreadyStarted(mod.Name())
	} else if err, mod.serialPort = mod.StringParam("gps.device"); err != nil {
		return err
	} else if err, mod.baudRate = mod.IntParam("gps.baudrate"); err != nil {
		return err
	}

	mod.serial, err = serial.OpenPort(&serial.Config{
		Name:        mod.serialPort,
		Baud:        mod.baudRate,
		ReadTimeout: time.Second * 1,
	})

	return
}

func (mod *GPS) readLine() (line string, err error) {
	var n int

	b := make([]byte, 1)
	for {
		if n, err = mod.serial.Read(b); err != nil {
			return
		} else if n == 1 {
			if b[0] == '\n' {
				return str.Trim(line), nil
			} else {
				line += string(b[0])
			}
		}
	}
}

func (mod *GPS) Show() error {
	fmt.Printf("latitude:%f longitude:%f quality:%s satellites:%d altitude:%f\n",
		mod.Session.GPS.Latitude,
		mod.Session.GPS.Longitude,
		mod.Session.GPS.FixQuality,
		mod.Session.GPS.NumSatellites,
		mod.Session.GPS.Altitude)

	mod.Session.Refresh()

	return nil
}

func (mod *GPS) Start() error {
	if err := mod.Configure(); err != nil {
		return err
	}

	return mod.SetRunning(true, func() {
		defer mod.serial.Close()

		mod.Info("started on port %s ...", mod.serialPort)

		for mod.Running() {
			if line, err := mod.readLine(); err == nil {
				if s, err := nmea.Parse(line); err == nil {
					// http://aprs.gids.nl/nmea/#gga
					if m, ok := s.(nmea.GNGGA); ok {
						mod.Session.GPS.Updated = time.Now()
						mod.Session.GPS.Latitude = m.Latitude
						mod.Session.GPS.Longitude = m.Longitude
						mod.Session.GPS.FixQuality = m.FixQuality
						mod.Session.GPS.NumSatellites = m.NumSatellites
						mod.Session.GPS.HDOP = m.HDOP
						mod.Session.GPS.Altitude = m.Altitude
						mod.Session.GPS.Separation = m.Separation
					} else if m, ok := s.(nmea.GPGGA); ok {
						mod.Session.GPS.Updated = time.Now()
						mod.Session.GPS.Latitude = m.Latitude
						mod.Session.GPS.Longitude = m.Longitude
						mod.Session.GPS.FixQuality = m.FixQuality
						mod.Session.GPS.NumSatellites = m.NumSatellites
						mod.Session.GPS.HDOP = m.HDOP
						mod.Session.GPS.Altitude = m.Altitude
						mod.Session.GPS.Separation = m.Separation
					}
				} else {
					mod.Debug("error parsing line '%s': %s", line, err)
				}
			} else if err != io.EOF {
				mod.Warning("error while reading serial port: %s", err)
			}
		}
	})
}

func (mod *GPS) Stop() error {
	return mod.SetRunning(false, func() {
		// let the read fail and exit
		mod.serial.Close()
	})
}
