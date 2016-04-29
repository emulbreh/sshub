package libsshub

type Tunnel struct {
	From        Port   `json:"from"`
	To          Port   `json:"to"`
	Port        uint32 `json:"port"`
	Established bool   `json:"-"`
}
