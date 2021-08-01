package shared

type Request struct {
	Address string `json:"address"`
	Tunnel  string `json:"string"`
	Key     string `json:"key"`
	User    string `json:"user"`
}
