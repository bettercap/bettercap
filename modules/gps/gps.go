package gps

import (
	"fmt"
	"io"
	"time"

	"github.com/bettercap/bettercap/session"

	"github.com/adrianmo/go-nmea"
	"github.com/stratoberry/go-gpsd"
	"github.com/tarm/serial"

	"github.com/evilsocket/islazy/str"
)

type GPS struct {
	session.SessionModule

	serialPort string
	baudRate   int

	serial *serial.Port
	gpsd   *gpsd.Session
}

var ModeInfo = [4]string{
	"NoValueSeen",
	"NoFix",
	"Mode2D",
	"Mode3D",
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
		"Serial device of the GPS hardware or hostname:port for a GPSD instance."))

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
	return "A module talking with GPS hardware on a serial interface or via GPSD."
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

	if mod.serialPort[0] == '/' || mod.serialPort[0] == '.' {
		mod.Debug("connecting to serial port %s", mod.serialPort)
		mod.serial, err = serial.OpenPort(&serial.Config{
			Name:        mod.serialPort,
			Baud:        mod.baudRate,
			ReadTimeout: time.Second * 1,
		})
	} else {
		mod.Debug("connecting to gpsd at %s", mod.serialPort)
		mod.gpsd, err = gpsd.Dial(mod.serialPort)
	}

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
	mod.Printf("latitude:%f longitude:%f quality:%s satellites:%d altitude:%f\n",
		mod.Session.GPS.Latitude,
		mod.Session.GPS.Longitude,
		mod.Session.GPS.FixQuality,
		mod.Session.GPS.NumSatellites,
		mod.Session.GPS.Altitude)

	mod.Session.Refresh()

	return nil
}

func (mod *GPS) readFromSerial() {
	if line, err := mod.readLine(); err == nil {
		if s, err := nmea.Parse(line); err == nil {
			// http://aprs.gids.nl/nmea/#gga
			if m, ok := s.(nmea.GGA); ok {
				mod.Session.GPS.Updated = time.Now()
				mod.Session.GPS.Latitude = m.Latitude
				mod.Session.GPS.Longitude = m.Longitude
				mod.Session.GPS.FixQuality = m.FixQuality
				mod.Session.GPS.NumSatellites = m.NumSatellites
				mod.Session.GPS.HDOP = m.HDOP
				mod.Session.GPS.Altitude = m.Altitude
				mod.Session.GPS.Separation = m.Separation

				mod.Session.Events.Add("gps.new", mod.Session.GPS)
			}
		} else {
			mod.Debug("error parsing line '%s': %s", line, err)
		}
	} else if err != io.EOF {
		mod.Warning("error while reading serial port: %s", err)
	}
}

func (mod *GPS) runFromGPSD() {
	mod.gpsd.AddFilter("TPV", func(r interface{}) {
		report := r.(*gpsd.TPVReport)
		mod.Session.GPS.Updated = report.Time
		mod.Session.GPS.Latitude = report.Lat
		mod.Session.GPS.Longitude = report.Lon
		mod.Session.GPS.FixQuality = ModeInfo[report.Mode]
		mod.Session.GPS.Altitude = report.Alt

		mod.Session.Events.Add("gps.new", mod.Session.GPS)
	})

	mod.gpsd.AddFilter("SKY", func(r interface{}) {
		report := r.(*gpsd.SKYReport)
		mod.Session.GPS.NumSatellites = int64(len(report.Satellites))
		mod.Session.GPS.HDOP = report.Hdop
		//mod.Session.GPS.Separation = 0
	})

	done := mod.gpsd.Watch()
	<-done
}

func (mod *GPS) Start() error {
	if err := mod.Configure(); err != nil {
		return err
	}

	return mod.SetRunning(true, func() {
		mod.Info("started on port %s ...", mod.serialPort)

		if mod.serial != nil {
			defer mod.serial.Close()

			for mod.Running() {
				mod.readFromSerial()
			}
		} else {
			mod.runFromGPSD()
		}
	})
}

func (mod *GPS) Stop() error {
	return mod.SetRunning(false, func() {
		if mod.serial != nil {
			// let the read fail and exit
			mod.serial.Close()
		} /*
			FIXME: no Close or Stop method in github.com/stratoberry/go-gpsd
			else {
				if err := mod.gpsd.Close(); err != nil {
					mod.Error("failed closing the connection to GPSD: %s", err)
				}
			} */
	})
}
