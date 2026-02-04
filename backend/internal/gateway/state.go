package gateway

type State int

const (
	StateDisconnected State = iota
	StateConnecting
	StateConnected
	StateDisconnecting
	StateDisconnectionError
)

func (s State) String() string {
	switch s {
	case StateDisconnected:
		return "disconnected"
	case StateConnecting:
		return "connecting"
	case StateConnected:
		return "connected"
	case StateDisconnecting:
		return "disconnecting"
	case StateDisconnectionError:
		return "disconnection error"
	default:
		return "unknown"
	}
}
