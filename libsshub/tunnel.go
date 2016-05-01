package libsshub

type Link struct {
	From        Tunnel `json:"from"`
	To          Tunnel `json:"to"`
	Port        uint32 `json:"port"`
	Established bool   `json:"-"`
}
