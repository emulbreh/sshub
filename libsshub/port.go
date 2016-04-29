package libsshub

import (
	"fmt"
	"golang.org/x/crypto/ssh"
)

type Port struct {
	PublicKey string   `json:"public_key"`
	User      string   `json:"user"`
	Tunnel    *Tunnel  `json:"-"`
	Conn      ssh.Conn `json:"-"`
}

type PortStatus struct {
	*Port
	Connected bool `json:"connected"`
}

func (port *Port) IsSource() bool {
	return port.Tunnel.From == *port
}

func (port *Port) Serialize() PortStatus {
	return PortStatus{port, port.Conn != nil}
}

func (port *Port) String() string {
	return fmt.Sprintf("Port(user=%s)", port.User)
}
