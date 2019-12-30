package gpsdaemon

import (
	"fmt"
	"github.com/bettercap/bettercap/session"

	"github.com/koppacetic/go-gpsd"
)

type GPSDaemon struct {
	session.SessionModule

	host             string
	port             int
	gpsdaemonSession *gpsd.Session
}

var ModeInfo = [4]string{"NoValueSeen", "NoFix", "Mode2D", "Mode3D"}

func NewGPSDaemon(s *session.Session) *GPSDaemon {
	mod := &GPSDaemon{
		SessionModule: session.NewSessionModule("gpsdaemon", s),
		host:          "localhost",
		port:          2947,
	}

	mod.AddParam(session.NewStringParameter("gpsdaemon.host",
		mod.host,
		"",
		"Hostname or IP of GPSD."))

	mod.AddParam(session.NewIntParameter("gpsdaemon.port",
		fmt.Sprintf("%d", mod.port),
		"TCP Port of GPSD."))

	mod.AddHandler(session.NewModuleHandler("gpsdaemon on", "",
		"Start acquiring from GPSD.",
		func(args []string) error {
			return mod.Start()
		}))

	mod.AddHandler(session.NewModuleHandler("gpsdaemon off", "",
		"Stop acquiring from GPSD.",
		func(args []string) error {
			return mod.Stop()
		}))

	mod.AddHandler(session.NewModuleHandler("gpsdaemon.show", "",
		"Show the last coordinates returned by GPSD.",
		func(args []string) error {
			return mod.Show()
		}))

	return mod
}

func (mod *GPSDaemon) Name() string {
	return "gpsdaemon"
}

func (mod *GPSDaemon) Description() string {
	return "A module talking with GPSD."
}

func (mod *GPSDaemon) Author() string {
	return "fheylis (github.com/fheylis)"
}

func (mod *GPSDaemon) Configure() (err error) {
	if mod.Running() {
		return session.ErrAlreadyStarted(mod.Name())
	} else if err, mod.host = mod.StringParam("gpsdaemon.host"); err != nil {
		return err
	} else if err, mod.port = mod.IntParam("gpsdaemon.port"); err != nil {
		return err
	}

	if mod.gpsdaemonSession, err = gpsd.Dial(fmt.Sprintf("%s:%d", mod.host, mod.port)); err != nil {
		mod.Error("Failed to connect to GPSD: %s", err)
	}

	return
}

func (mod *GPSDaemon) Show() error {
	fmt.Printf("latitude:%f longitude:%f quality:%s satellites:%d altitude:%f\n",
		mod.Session.GPS.Latitude,
		mod.Session.GPS.Longitude,
		mod.Session.GPS.FixQuality,
		mod.Session.GPS.NumSatellites,
		mod.Session.GPS.Altitude)

	mod.Session.Refresh()

	return nil
}

func (mod *GPSDaemon) Start() error {
	if err := mod.Configure(); err != nil {
		return err
	}

	return mod.SetRunning(true, func() {
		mod.Info("started on GPSD %s:%d ...", mod.host, mod.port)

		mod.gpsdaemonSession.Subscribe("TPV", func(r interface{}) {
			report := r.(*gpsd.TPVReport)

			mod.Session.GPS.Updated = report.Time
			mod.Session.GPS.Latitude = report.Lat
			mod.Session.GPS.Longitude = report.Lon
			mod.Session.GPS.FixQuality = ModeInfo[report.Mode]
			mod.Session.GPS.Altitude = report.Alt
		})

		mod.gpsdaemonSession.Subscribe("SKY", func(r interface{}) {
			report := r.(*gpsd.SKYReport)

			mod.Session.GPS.NumSatellites = int64(len(report.Satellites))
			mod.Session.GPS.HDOP = report.Hdop
			//mod.Session.GPS.Separation = 0
		})

		mod.gpsdaemonSession.Run()
	})
}

func (mod *GPSDaemon) Stop() error {
	return mod.SetRunning(false, func() {
		if err := mod.gpsdaemonSession.Close(); err != nil {
			mod.Error("Failed closing the connection to GPSD: %s", err)
		}
	})
}
