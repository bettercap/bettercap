package routing

type RouteType int

const (
	IPv4 RouteType = 0
	IPv6 RouteType = 1
)

type Route struct {
	Type        RouteType
	Default     bool
	Device      string
	Destination string
	Gateway     string
	Flags       string
}
