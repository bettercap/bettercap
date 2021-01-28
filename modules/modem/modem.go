/*
# Bettercap Serial Modem Module

### Preface :
@evilsocket recently added the C2 module and I thought "Hey! wouldn't be fun (and socially irresponsible)
if we could control bettercap over a 3g/4g network connection? Well friends here ya go!

This module depends on modemmanager. Some USB modems require mode-switching to behave as a serial
interface - if this is you (Huaweii and others) you'll need usb_modeswitch. The mod is very rough
around the edges and could definitely do for a refactor by someone good who actually knows what
they are doing.

### What works :

* Configure APN for network connectivity
* Display some groovy modem info
* Send / receive SMS message
* Read / clear SMS message from SIM


### What hasn't been tested?

* Every modem that is not the Huaweii E303C
* When things fail like if SIM is missing or cannot register on network or pretty much everything


### Future things?

* C2 over SMS (like the IRC module except different)
* Templating modem output for whatever reason


### What kinda data does this provide?

A lot, some of it very cool like current cell tower and operator, ICCID, IMEI, tons of stuff,.
Here's what mmcli kinda looks like with some fields redacted :

  -----------------------------
  General  |         dbus path: /org/freedesktop/ModemManager1/Modem/1
           |         device id: 0000000000000000000000000000000000000000
  -----------------------------
  Hardware |      manufacturer: huawei
           |             model: E303C
           | firmware revision: 21.157.01.01.18
           |         supported: gsm-umts
           |           current: gsm-umts
           |      equipment id: 000000000000000
  -----------------------------
  System   |            device: /sys/devices/pci0000:00/0000:00:1d.0/usb2/2-1/2-1.3/2-1.3.4/2-1.3.4.2
           |           drivers: option1
           |            plugin: Huawei
           |      primary port: ttyUSB2
           |             ports: ttyUSB2 (at), ttyUSB3 (at)
  -----------------------------
  Status   |    unlock retries: sim-pin (5), sim-puk (10), sim-pin2 (5), sim-puk2 (10)
           |             state: connected
           |       power state: on
           |       access tech: umts
           |    signal quality: 58% (recent)
  -----------------------------
  Modes    |         supported: allowed: 2g, 3g; preferred: none
           |                    allowed: 2g, 3g; preferred: 2g
           |                    allowed: 2g, 3g; preferred: 3g
           |                    allowed: 2g; preferred: none
           |                    allowed: 3g; preferred: none
           |           current: allowed: 2g, 3g; preferred: 3g
  -----------------------------
  IP       |         supported: ipv4, ipv6, ipv4v6
  -----------------------------
  3GPP     |              imei: 000000000000000
           |       operator id: 000000
           |     operator name: Stupid Expensive Mobile Carrier Inc.
           |      registration: home
  -----------------------------
  SIM      |         dbus path: /org/freedesktop/ModemManager1/SIM/1
  -----------------------------
  Bearer   |         dbus path: /org/freedesktop/ModemManager1/Bearer/2


### Dragons! Everywhere!

This module does NOT connect you to the internet - it simply registers the modem with the carrier.
If your modem appears as /dev/cdc-* a network interface will be created by the kernel and you can
request DHCP. If your modem inerface looks like /dev/ttyUSB* I have no idea if this will work, it will
probably use PPP or something, dunno - try it out send me a report or whatever :)

Despite having a config option for PIN it doesn't do anything lol; the SIM's I've tested do not
need any of that junk but if yours does well I'm sorry but I'm a loner Dotty, a rebel.

I don't think this module should be responsible for configuring your network iface but maybe I'm wrong.
If you think I am wrong ask the boss @evilsocket and mabs he'll do a thing about it.

## This module should be considered very unstable && I am NOT responsible for your phone bill ##

You have been warned! <3
*/

package modem

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/bettercap/bettercap/modules/events_stream"
	"github.com/bettercap/bettercap/session"
	"github.com/evilsocket/islazy/log"
)

type settings struct {
	device string
	apn    string
	pin    string
}

var mmcliPath string

type Modem struct {
	session.SessionModule

	settings settings
	stream   *events_stream.EventsStream
	eventBus session.EventBus
	quit     chan bool

	Modem    *MMCLIModem
	Location *ModemLocation
	Messages map[string]*SMS
	SIM      *SIM
	Bearer   *Bearer
}

// MMCLIModem converts the output of mmclie -m {modem} from JSON to struct
type MMCLIModem struct {
	Modem struct {
		ThreeGpp struct {
			EnabledLocks []interface{} `json:"enabled-locks"`
			Eps          struct {
				InitialBearer struct {
					DbusPath string `json:"dbus-path"`
					Settings struct {
						Apn      string `json:"apn"`
						IPType   string `json:"ip-type"`
						Password string `json:"password"`
						User     string `json:"user"`
					} `json:"settings"`
				} `json:"initial-bearer"`
				UeModeOperation string `json:"ue-mode-operation"`
			} `json:"eps"`
			Imei              string `json:"imei"`
			OperatorCode      string `json:"operator-code"`
			OperatorName      string `json:"operator-name"`
			Pco               string `json:"pco"`
			RegistrationState string `json:"registration-state"`
		} `json:"3gpp"`
		Cdma struct {
			ActivationState         string `json:"activation-state"`
			Cdma1XRegistrationState string `json:"cdma1x-registration-state"`
			Esn                     string `json:"esn"`
			EvdoRegistrationState   string `json:"evdo-registration-state"`
			Meid                    string `json:"meid"`
			Nid                     string `json:"nid"`
			Sid                     string `json:"sid"`
		} `json:"cdma"`
		DbusPath string `json:"dbus-path"`
		Generic  struct {
			AccessTechnologies           []interface{} `json:"access-technologies"`
			Bearers                      []interface{} `json:"bearers"`
			CarrierConfiguration         string        `json:"carrier-configuration"`
			CarrierConfigurationRevision string        `json:"carrier-configuration-revision"`
			CurrentBands                 []interface{} `json:"current-bands"`
			CurrentCapabilities          []string      `json:"current-capabilities"`
			CurrentModes                 string        `json:"current-modes"`
			Device                       string        `json:"device"`
			DeviceIdentifier             string        `json:"device-identifier"`
			Drivers                      []string      `json:"drivers"`
			EquipmentIdentifier          string        `json:"equipment-identifier"`
			HardwareRevision             string        `json:"hardware-revision"`
			Manufacturer                 string        `json:"manufacturer"`
			Model                        string        `json:"model"`
			OwnNumbers                   []interface{} `json:"own-numbers"`
			Plugin                       string        `json:"plugin"`
			Ports                        []string      `json:"ports"`
			PowerState                   string        `json:"power-state"`
			PrimaryPort                  string        `json:"primary-port"`
			Revision                     string        `json:"revision"`
			SignalQuality                struct {
				Recent string `json:"recent"`
				Value  string `json:"value"`
			} `json:"signal-quality"`
			Sim                   string        `json:"sim"`
			State                 string        `json:"state"`
			StateFailedReason     string        `json:"state-failed-reason"`
			SupportedBands        []interface{} `json:"supported-bands"`
			SupportedCapabilities []string      `json:"supported-capabilities"`
			SupportedIPFamilies   []string      `json:"supported-ip-families"`
			SupportedModes        []string      `json:"supported-modes"`
			UnlockRequired        string        `json:"unlock-required"`
			UnlockRetries         []string      `json:"unlock-retries"`
		} `json:"generic"`
	} `json:"modem"`
}

type eventContext struct {
	Session *session.Session
	Event   session.Event
}

// SIM converts the output of mmcli --sim {path} from JSON into a struct
type SIM struct {
	Sim struct {
		DbusPath   string `json:"dbus-path"`
		Properties struct {
			EmergencyNumbers []interface{} `json:"emergency-numbers"`
			Iccid            string        `json:"iccid"`
			Imsi             string        `json:"imsi"`
			OperatorCode     string        `json:"operator-code"`
			OperatorName     string        `json:"operator-name"`
		} `json:"properties"`
	} `json:"sim"`
}

// ModemList converts the output of mmcli -L from JSON into a struct
type ModemList struct {
	ModemList []string `json:"modem-list"`
}

// SMSList is a list of SMS messages as strings
type SMSList struct {
	ModemMessagingSms []string `json:"modem.messaging.sms"`
}

// SMS converts the output of mmcli -m {modem} -s {SIM} from JSON to struct
type SMS struct {
	Sms struct {
		Content struct {
			Data   string `json:"data"`
			Number string `json:"number"`
			Text   string `json:"text"`
		} `json:"content"`
		DbusPath   string `json:"dbus-path"`
		Properties struct {
			Class              string `json:"class"`
			DeliveryReport     string `json:"delivery-report"`
			DeliveryState      string `json:"delivery-state"`
			DischargeTimestamp string `json:"discharge-timestamp"`
			MessageReference   string `json:"message-reference"`
			PduType            string `json:"pdu-type"`
			ServiceCategory    string `json:"service-category"`
			Smsc               string `json:"smsc"`
			State              string `json:"state"`
			Storage            string `json:"storage"`
			TeleserviceID      string `json:"teleservice-id"`
			Timestamp          string `json:"timestamp"`
			Validity           string `json:"validity"`
		} `json:"properties"`
	} `json:"sms"`
}

// ModemLocation converts the output of mmcli --location-get from JSON to struct
type ModemLocation struct {
	Modem struct {
		Location struct {
			ThreeGpp struct {
				Cid string `json:"cid"`
				Lac string `json:"lac"`
				Mcc string `json:"mcc"`
				Mnc string `json:"mnc"`
				Tac string `json:"tac"`
			} `json:"3gpp"`
			CdmaBs struct {
				Latitude  string `json:"latitude"`
				Longitude string `json:"longitude"`
			} `json:"cdma-bs"`
			Gps struct {
				Altitude  string        `json:"altitude"`
				Latitude  string        `json:"latitude"`
				Longitude string        `json:"longitude"`
				Nmea      []interface{} `json:"nmea"`
				Utc       string        `json:"utc"`
			} `json:"gps"`
		} `json:"location"`
	} `json:"modem"`
}

// Bearer converts the output of mmcli -b {bearer} from JSON to struct
type Bearer struct {
	Bearer struct {
		DbusPath   string `json:"dbus-path"`
		Ipv4Config struct {
			Address string   `json:"address"`
			DNS     []string `json:"dns"`
			Gateway string   `json:"gateway"`
			Method  string   `json:"method"`
			Mtu     string   `json:"mtu"`
			Prefix  string   `json:"prefix"`
		} `json:"ipv4-config"`
		Ipv6Config struct {
			Address string        `json:"address"`
			DNS     []interface{} `json:"dns"`
			Gateway string        `json:"gateway"`
			Method  string        `json:"method"`
			Mtu     string        `json:"mtu"`
			Prefix  string        `json:"prefix"`
		} `json:"ipv6-config"`
		Properties struct {
			AllowedAuth []interface{} `json:"allowed-auth"`
			Apn         string        `json:"apn"`
			IPType      string        `json:"ip-type"`
			Password    string        `json:"password"`
			RmProtocol  string        `json:"rm-protocol"`
			Roaming     string        `json:"roaming"`
			User        string        `json:"user"`
		} `json:"properties"`
		Stats struct {
			BytesRx  string `json:"bytes-rx"`
			BytesTx  string `json:"bytes-tx"`
			Duration string `json:"duration"`
		} `json:"stats"`
		Status struct {
			Connected string `json:"connected"`
			Interface string `json:"interface"`
			IPTimeout string `json:"ip-timeout"`
			Suspended string `json:"suspended"`
		} `json:"status"`
		Type string `json:"type"`
	} `json:"bearer"`
}

// NewModem makes a shiny new modem
func NewModem(s *session.Session) *Modem {

	mod := &Modem{
		SessionModule: session.NewSessionModule("modem", s),
		stream:        events_stream.NewEventsStream(s),
		quit:          make(chan bool),
		settings:      settings{},
		Messages:      make(map[string]*SMS),
	}

	mod.AddParam(session.NewStringParameter("modem.apn",
		mod.settings.apn,
		"",
		"APN"))

	mod.AddParam(session.NewStringParameter("modem.pin",
		mod.settings.pin,
		"",
		"PIN for SIM card"))

	mod.AddHandler(session.NewModuleHandler("modem on",
		"",
		"Start the modem module.",
		func(args []string) error {
			return mod.Start()
		}))

	mod.AddHandler(session.NewModuleHandler("modem off",
		"",
		"Stop the modem module.",
		func(args []string) error {
			return mod.Stop()
		}))

	mod.AddHandler(session.NewModuleHandler("modem.show", "",
		"Display modem connection and location information.",
		func(args []string) error {
			return mod.Show()
		}))

	mod.AddHandler(session.NewModuleHandler("modem.sms.show", "",
		"Display SMS messages.",
		func(args []string) error {
			return mod.ReadSMS()
		}))

	mod.AddHandler(session.NewModuleHandler("modem.sms.clear", "",
		"Clear/delete SMS messages.",
		func(args []string) error {
			return mod.ClearSMS()
		}))

	mod.AddHandler(session.NewModuleHandler("modem.sms.send NUMBER MESSAGE",
		"modem.sms.send ([^\\s]+) (.+)",
		"Send an SMS message.",
		func(args []string) error {
			var number, message string
			if ok := args[0]; ok != "" {
				number = ok
			}
			if ok := args[1]; ok != "" {
				message = strings.Join(args[1:], " ")
			}
			mod.SendSMS(number, message)
			mod.Debug("SMS sent to %s", number)
			return nil
		}))

	return mod
}

func (mod *Modem) Name() string {
	return "modem"
}

func (mod *Modem) Description() string {
	return "A ModemManager module that can send and receive SMS messages, connect to 3g/4g carrier and display connection info."
}

func (mod *Modem) Author() string {
	return "https://github.com/aster1sk"
}

func (mod *Modem) Configure() (err error) {

	if mod.Running() {
		return session.ErrAlreadyStarted(mod.Name())
	}

	if err, mod.settings.device = mod.StringParam("modem.apn"); err != nil {
		return err
	} else if err, mod.settings.pin = mod.StringParam("modem.pin"); err != nil {
		return err
	}

	mod.eventBus = mod.Session.Events.Listen()

	if log.Level == log.DEBUG {
		// @TODO maybe make this work
	}

	return err
}

func (mod *Modem) onEvent(e session.Event) {

	if mod.Session.EventsIgnoreList.Ignored(e) {
		mod.Info("%v", e.Data)
		return
	}

	return
}

func (mod *Modem) Start() error {

	if err := mod.Configure(); err != nil {
		return err
	}

	// This kinda feels against bettercap conventions
	go mod.ReadModemData()

	return mod.SetRunning(true, func() {
		for mod.Running() {
			var e session.Event
			select {
			case e = <-mod.eventBus:
				mod.onEvent(e)
			case <-mod.quit:
				mod.Debug("got quit")
				return
			}
		}
	})
}

// Stop stops the modem module
func (mod *Modem) Stop() error {
	mod.Info("Stopping modem %s", mod.settings.device)
	// @TODO mmcli disable?
	return mod.SetRunning(false, func() {
		mod.quit <- true
		mod.Session.Events.Unlisten(mod.eventBus)
	})
}

// Show displays information about the modem :
func (mod *Modem) Show() error {
	// @TODO tabular output is ideal here like the wifi.show functionality, not sure how
	// There is a lot of cool / interesting stuff in here like cell tower ID etc, perhaps this should be templated?
	mod.Printf("Registered : '%s' Carrier : '%s' ICCID : %s RSSI %s%% \n",
		mod.Modem.Modem.ThreeGpp.RegistrationState,
		mod.Modem.Modem.ThreeGpp.OperatorName,
		mod.SIM.Sim.Properties.Iccid,
		mod.Modem.Modem.Generic.SignalQuality.Value,
	)
	return nil
}

// SendSMS sends an SMS to {number} with {message}
func (mod *Modem) SendSMS(number, message string) error {

	if number == "" || message == "" {
		return fmt.Errorf("number and message required")
	}

	// @TODO shell escape message and number else pwnd!
	cmd := fmt.Sprintf(`%s -m %s --messaging-create-sms=text="%s",number="%s"`,
		mmcliPath,
		mod.Modem.Modem.DbusPath,
		message,
		number,
	)

	mod.Info(runCommand(cmd))

	for s, x := range mod.Messages {
		m := mod.Modem.Modem.DbusPath
		if x.Sms.Properties.State == "--" {
			/* I'm pretty sure this works universally, definitely requires further testing.
			Send the message if it's not 'sent' or 'received' state.
			This probably needs some better error handling */
			mod.Info(runCommand(fmt.Sprintf("%s -m %s -s %s --send", mmcliPath, m, s)))
		}
	}

	return nil
}

// ReadSMS displays sent and received SMS messages
func (mod *Modem) ReadSMS() error {
	// @TODO tabular output is ideal here like the wifi.show functionality, not sure how
	for _, msg := range mod.Messages {
		mod.Info("from %s : %s (%s)", msg.Sms.Content.Number, msg.Sms.Content.Text, msg.Sms.Properties.State)
	}
	return nil
}

// ClearSMS will purge all messages from SIM and module
func (mod *Modem) ClearSMS() error {
	for s := range mod.Messages {
		m := mod.Modem.Modem.DbusPath
		mod.Info(runCommand(fmt.Sprintf("%s -m %s --messaging-delete-sms=%s", mmcliPath, m, s)))
		delete(mod.Messages, s)
	}
	return nil
}

// ReadModemData reads data from the modem
func (mod *Modem) ReadModemData() {

	if p, ok := exec.LookPath("mmcli"); ok == nil {
		mmcliPath = p
	} else {
		// Is the the correct way to do this? This is a panic() case
		mod.Error("Could not find mmcli")
		mod.Stop()
		return
	}

	// Probably best to use the built-in ticker from bettercap but that REEKS of effort.
	for range time.Tick(time.Second * 1) {

		// This is one gnarly long function that should probably be split up into different parts
		out := runCommand(mmcliPath + " -L -J")
		var modemList ModemList
		json.Unmarshal([]byte(out), &modemList)

		for _, m := range modemList.ModemList {

			out = runCommand(fmt.Sprintf("%s -m %s -J", mmcliPath, m))
			var modem MMCLIModem
			json.Unmarshal([]byte(out), &modem)
			mod.settings.device = modem.Modem.DbusPath

			switch modem.Modem.Generic.State {

			case "disabled":
				runCommand(fmt.Sprintf("%s -m %s -e", mmcliPath, m))

			case "registered":
				operCode := modem.Modem.ThreeGpp.OperatorCode
				_ = operCode
				cmd := fmt.Sprintf(`%s -m %s --simple-connect=apn=%s`, mmcliPath, m, mod.settings.apn)
				runCommand(cmd)

			case "connected":
				bearers := modem.Modem.Generic.Bearers
				runCommand(fmt.Sprintf("%s --modem-set-enable-signal -m %s -J", mmcliPath, m))
				for _, b := range bearers {
					cmd1 := fmt.Sprintf("%s -b %s -J", mmcliPath, b)
					out = runCommand(cmd1)
					var bearer Bearer
					json.Unmarshal([]byte(out), &bearer)

					i := 1
					for _, d := range bearer.Bearer.Ipv4Config.DNS {
						dnsTxt := fmt.Sprintf("    DNS Server %d", i)
						_ = d
						_ = dnsTxt
						i++
					}

					s := modem.Modem.Generic.Sim
					out = runCommand(fmt.Sprintf("%s --sim %s -J", mmcliPath, s))
					var sim SIM
					json.Unmarshal([]byte(out), &sim)

					out = runCommand(fmt.Sprintf("%s --location-get -m %s -J", mmcliPath, m))
					var location ModemLocation
					json.Unmarshal([]byte(out), &location)

					out = runCommand(fmt.Sprintf("%s -m %s --messaging-list-sms -J", mmcliPath, m))
					var smsl SMSList
					json.Unmarshal([]byte(out), &smsl)

					smss := make(map[string]*SMS)

					var msgCount int
					for _, s := range smsl.ModemMessagingSms {
						msgCount++
						out = runCommand(fmt.Sprintf("%s -m %s -s %s -J", mmcliPath, m, s))
						var sms SMS
						json.Unmarshal([]byte(out), &sms)
						smss[s] = &sms
					}
					if len(mod.Messages) != msgCount {
						mod.ReadSMS()
					}

					// @TODO likely a buttload of race conditions happening here ¯\_(ツ)_/¯
					mod.Messages = smss
					mod.Location = &location
					mod.SIM = &sim
					mod.Bearer = &bearer
					mod.Modem = &modem

					/* Maybe these events should be configurable for example :
					if rssi != previous rssi -> print msg
					*/
					if mod.Modem.Modem.Generic.Device != modem.Modem.Generic.Device {
						evt := session.NewEvent("modem.device", mod.Modem)
						mod.Info("%v", evt)
					}
				}
			}
		}
	}
}

// runCommand is a terrible thing and should be replaced with the better one that bettercap uses for !eval
func runCommand(command string) string {
	parts := strings.Fields(command)
	baseCmd := parts[0]
	cmdArgs := parts[1:]
	cmd := exec.Command(baseCmd, cmdArgs...)
	r, _ := cmd.StdoutPipe()
	cmd.Stderr = cmd.Stdout
	scanner := bufio.NewScanner(r)
	err := cmd.Start()
	if err != nil {
		panic(err)
	}
	var output string
	for scanner.Scan() {
		output += scanner.Text()
	}
	cmd.Wait()
	return output
}

// superflous comment for sloc win
