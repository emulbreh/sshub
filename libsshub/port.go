package libsshub

import (
	"fmt"
	"golang.org/x/crypto/ssh"
)

type Tunnel struct {
	PublicKey string   `json:"public_key"`
	User      string   `json:"user"`
	Link      *Link    `json:"-"`
	Conn      ssh.Conn `json:"-"`
}

type TunnelStatus struct {
	*Tunnel
	Connected bool `json:"connected"`
}

func (tun *Tunnel) IsSource() bool {
	return tun.Link.From == *tun
}

func (tun *Tunnel) Serialize() TunnelStatus {
	return TunnelStatus{tun, tun.Conn != nil}
}

func (tun *Tunnel) String() string {
	return fmt.Sprintf("Port(user=%s)", tun.User)
}
