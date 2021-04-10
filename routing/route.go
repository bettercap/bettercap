package routing

type RouteType string

const (
	IPv4 RouteType = "IPv4"
	IPv6 RouteType = "IPv6"
)

type Route struct {
	Type        RouteType
	Default     bool
	Device      string
	Destination string
	Gateway     string
	Flags       string
}
