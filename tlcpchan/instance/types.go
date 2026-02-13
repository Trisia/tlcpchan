package instance

type Status string

const (
	StatusCreated Status = "created"
	StatusRunning Status = "running"
	StatusStopped Status = "stopped"
	StatusError   Status = "error"
)

type InstanceType string

const (
	TypeServer     InstanceType = "server"
	TypeClient     InstanceType = "client"
	TypeHTTPServer InstanceType = "http-server"
	TypeHTTPClient InstanceType = "http-client"
)

func ParseInstanceType(s string) InstanceType {
	switch s {
	case "server":
		return TypeServer
	case "client":
		return TypeClient
	case "http-server":
		return TypeHTTPServer
	case "http-client":
		return TypeHTTPClient
	default:
		return TypeServer
	}
}
