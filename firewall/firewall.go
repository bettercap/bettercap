package firewall

type FirewallManager interface {
	IsForwardingEnabled() bool
	EnableForwarding(enabled bool) error
	EnableRedirection(r *Redirection, enabled bool) error
	Restore()
}
