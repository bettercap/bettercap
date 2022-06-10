package any_proxy

import (
	"fmt"
	"github.com/bettercap/bettercap/firewall"
	"github.com/bettercap/bettercap/session"
	"github.com/evilsocket/islazy/str"
	"strconv"
	"strings"
)

type AnyProxy struct {
	session.SessionModule
	// not using map[int]*firewall.Redirection to preserve order
	ports        []int
	redirections []*firewall.Redirection
}

func NewAnyProxy(s *session.Session) *AnyProxy {
	mod := &AnyProxy{
		SessionModule: session.NewSessionModule("any.proxy", s),
	}

	mod.AddParam(session.NewStringParameter("any.proxy.iface",
		session.ParamIfaceName,
		"",
		"Interface to redirect packets from."))

	mod.AddParam(session.NewStringParameter("any.proxy.protocol",
		"TCP",
		"(TCP|UDP)",
		"Proxy protocol."))

	mod.AddParam(session.NewStringParameter("any.proxy.src_port",
		"80",
		"",
		"Remote port to redirect when the module is activated, "+
			"also supported a comma separated list of ports and/or port-ranges."))

	mod.AddParam(session.NewStringParameter("any.proxy.src_address",
		"",
		"",
		"Leave empty to intercept any source address."))

	mod.AddParam(session.NewStringParameter("any.proxy.dst_address",
		session.ParamIfaceAddress,
		"",
		"Address where the proxy is listening."))

	mod.AddParam(session.NewIntParameter("any.proxy.dst_port",
		"8080",
		"Port where the proxy is listening."))

	mod.AddHandler(session.NewModuleHandler("any.proxy on", "",
		"Start the custom proxy redirection.",
		func(args []string) error {
			return mod.Start()
		}))

	mod.AddHandler(session.NewModuleHandler("any.proxy off", "",
		"Stop the custom proxy redirection.",
		func(args []string) error {
			return mod.Stop()
		}))

	return mod
}

func (mod *AnyProxy) Name() string {
	return "any.proxy"
}

func (mod *AnyProxy) Description() string {
	return "A firewall redirection to any custom proxy."
}

func (mod *AnyProxy) Author() string {
	return "Simone Margaritelli <evilsocket@gmail.com>"
}

func (mod *AnyProxy) Configure() error {
	var err error
	var srcPorts string
	var dstPort int
	var iface string
	var protocol string
	var srcAddress string
	var dstAddress string

	if mod.Running() {
		return session.ErrAlreadyStarted(mod.Name())
	} else if err, iface = mod.StringParam("any.proxy.iface"); err != nil {
		return err
	} else if err, protocol = mod.StringParam("any.proxy.protocol"); err != nil {
		return err
	} else if err, dstPort = mod.IntParam("any.proxy.dst_port"); err != nil {
		return err
	} else if err, srcAddress = mod.StringParam("any.proxy.src_address"); err != nil {
		return err
	} else if err, dstAddress = mod.StringParam("any.proxy.dst_address"); err != nil {
		return err
	}

	if err, srcPorts = mod.StringParam("any.proxy.src_port"); err != nil {
		return err
	} else {
		var ports []int
		// srcPorts can be a single port, a list of ports or a list of ranges, or a mix.
		for _, token := range str.Comma(str.Trim(srcPorts)) {
			if p, err := strconv.Atoi(token); err == nil {
				// simple case, integer port
				ports = append(ports, p)
			} else if strings.Contains(token, "-") {
				// port range
				if parts := strings.Split(token, "-"); len(parts) == 2 {
					if from, err := strconv.Atoi(str.Trim(parts[0])); err != nil {
						return fmt.Errorf("invalid start port %s: %s", parts[0], err)
					} else if from < 1 || from > 65535 {
						return fmt.Errorf("port %s out of valid range", parts[0])
					} else if to, err := strconv.Atoi(str.Trim(parts[1])); err != nil {
						return fmt.Errorf("invalid end port %s: %s", parts[1], err)
					} else if to < 1 || to > 65535 {
						return fmt.Errorf("port %s out of valid range", parts[1])
					} else if from > to {
						return fmt.Errorf("start port should be lower than end port")
					} else {
						for p := from; p <= to; p++ {
							ports = append(ports, p)
						}
					}
				} else {
					return fmt.Errorf("can't parse '%s' as range", token)
				}
			} else {
				return fmt.Errorf("can't parse '%s' as port or range", token)
			}
		}

		// after parsing and validation, create a redirection per source port
		mod.ports = ports
		mod.redirections = nil
		for _, port := range mod.ports {
			redir := firewall.NewRedirection(iface,
				protocol,
				port,
				dstAddress,
				dstPort)

			if srcAddress != "" {
				redir.SrcAddress = srcAddress
			}

			mod.redirections = append(mod.redirections, redir)
		}
	}

	if !mod.Session.Firewall.IsForwardingEnabled() {
		mod.Info("Enabling forwarding.")
		mod.Session.Firewall.EnableForwarding(true)
	}

	for _, redir := range mod.redirections {
		if err := mod.Session.Firewall.EnableRedirection(redir, true); err != nil {
			return err
		}
		mod.Info("applied redirection %s", redir.String())
	}

	return nil
}

func (mod *AnyProxy) Start() error {
	if err := mod.Configure(); err != nil {
		return err
	}

	return mod.SetRunning(true, func() {})
}

func (mod *AnyProxy) Stop() error {
	for _, redir := range mod.redirections {
		mod.Info("disabling redirection %s", redir.String())
		if err := mod.Session.Firewall.EnableRedirection(redir, false); err != nil {
			return err
		}
	}
	return mod.SetRunning(false, func() {})
}
