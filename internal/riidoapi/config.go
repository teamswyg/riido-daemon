package riidoapi

type Config struct {
	AppVersion string         `json:"app_version"`
	SocketPath string         `json:"socket_path"`
	TaskDBPath string         `json:"task_db_path"`
	Transport  LocalTransport `json:"transport"`
}

type Server struct {
	config Config
}

func NewServer(config Config) Server {
	return Server{config: config}
}
