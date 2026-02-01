package gateway

type State int

const (
	StateDisconnected State = iota
	StateDiscoveryConnecting
	StateDiscoveryConnected
	StateDataConnecting
	StateDataConnected
)

func (s State) String() string {
	switch s {
	case StateDisconnected:
		return "disconnected"
	case StateDiscoveryConnecting:
		return "connecting to LNS Discovery"
	case StateDiscoveryConnected:
		return "connected to LNS Discovery"
	case StateDataConnecting:
		return "connecting to LNS Data"
	case StateDataConnected:
		return "connected to LNS Data"
	default:
		return "unknown"
	}
}
