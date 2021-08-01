package shared

type Request struct {
	Host string `json:"address"`
	Tunnel  string `json:"string"`
	Key     string `json:"key"`
	User    string `json:"user"`
}
