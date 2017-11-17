package firewall

type FirewallManager interface {
	IsForwardingEnabled() bool
	EnableForwarding(enabled bool) error
	EnableIcmpBcast(enabled bool) error
	EnableSendRedirects(enabled bool) error
	EnableRedirection(r *Redirection, enabled bool) error
	Restore()
}
