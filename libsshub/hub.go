package libsshub

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"io"
	"net"
	"strings"
	"sync"
)

type Hub struct {
	mutex         sync.RWMutex
	links         []*Link
	tunnelsByUser map[string]*Tunnel
	privateKey    ssh.Signer
}

func NewHub(privateKey ssh.Signer) *Hub {
	return &Hub{
		mutex:         sync.RWMutex{},
		links:         []*Link{},
		tunnelsByUser: map[string]*Tunnel{},
		privateKey:    privateKey,
	}
}

func (hub *Hub) Listen(addr string) error {
	config := &ssh.ServerConfig{
		PublicKeyCallback: func(c ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			k := strings.TrimSpace(string(ssh.MarshalAuthorizedKey(key)))
			tunnel := hub.GetTunnelByUser(c.User())
			if tunnel == nil {
				log.Errorf("Authentication failed (unknown user): %s@%s", c.User(), c.RemoteAddr())
				return nil, fmt.Errorf("unknown user")
			}
			if k != tunnel.PublicKey {
				log.Errorf("Authentication failed (invalid key): %s@%s", c.User(), c.RemoteAddr())
				return nil, fmt.Errorf("unknown key")
			}
			log.Infof("Authentication successful: %s@%s", c.User(), c.RemoteAddr())
			log.Debugf("public key: %s %s", k, key.Type())
			return &ssh.Permissions{}, nil
		},
	}
	config.AddHostKey(hub.privateKey)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Errorf("Failed to setup tcp listener: %v", addr, err)
		return err
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Errorf("Failed to accept tcp connection: %v", err)
			continue
		}
		go hub.handleConnection(conn, config)
	}
	return nil
}

type forwardChannelArgs struct {
	Addr       string
	Port       uint32
	OriginAddr string
	OriginPort uint32
}

func (hub *Hub) addLink(link *Link) error {
	link.From.Link = link
	link.To.Link = link
	hub.mutex.Lock()
	defer hub.mutex.Unlock()
	if hub.tunnelsByUser[link.From.User] != nil {
		return fmt.Errorf("Multiple ports for user %s", link.From.User)
	}
	log.Infof("Configuring tunnel %s -> %s", link.From.User, link.To.User)
	hub.links = append(hub.links, link)
	hub.tunnelsByUser[link.From.User] = &link.From
	hub.tunnelsByUser[link.To.User] = &link.To
	return nil
}

func (hub *Hub) GetTunnelByUser(user string) *Tunnel {
	hub.mutex.RLock()
	defer hub.mutex.RUnlock()
	return hub.tunnelsByUser[user]
}

func DiscardChannels(chanReqs <-chan ssh.NewChannel) {
	for chanReq := range chanReqs {
		chanReq.Reject(ssh.Prohibited, "tcpip-forward only (-NR)")
	}
}

func (hub *Hub) handleRChannels(tunnel *Tunnel, conn ssh.Conn, chanReqs <-chan ssh.NewChannel) {
	for chanReq := range chanReqs {
		chanReq.Reject(ssh.Prohibited, "tcpip-forward only (-NR)")
	}
}

func (hub *Hub) handleRRequests(tunnel *Tunnel, conn ssh.Conn, reqs <-chan *ssh.Request) {
	for req := range reqs {
		if req.Type == "tcpip-forward" {
			tunnel.Conn = conn
			defer func() {
				tunnel.Conn = nil
			}()
			req.Reply(true, []byte{})
		} else {
			log.Warnf("got unexpected request %q WantReply=%q: %q\n", req.Type, req.WantReply, req.Payload)
			req.Reply(false, nil)
		}
	}
}

/*
func (hub *Hub) handleLChannels(chanReqs <-chan ssh.NewChannel) {

}

func (hub *Hub) handleLRequests(port Port, reqs <-ssh.Request) {
}
*/

func (hub *Hub) handleConnection(netConn net.Conn, config *ssh.ServerConfig) error {
	conn, chanReqs, reqs, err := ssh.NewServerConn(netConn, config)
	if err != nil {
		log.Errorf("Failed to initialize ssh connection from %v: %v", netConn.RemoteAddr(), err)
		return err
	}
	remoteAddr := conn.RemoteAddr().(*net.TCPAddr)
	log.Infof("Initialized ssh connection to %s@%v", conn.User(), remoteAddr)
	tunnel := hub.GetTunnelByUser(conn.User())
	if tunnel == nil {
		log.Errorf("No tunnel configuration for %v", conn.User())
	}

	if tunnel.IsSource() {
		go ssh.DiscardRequests(reqs)
		for chanReq := range chanReqs {
			if chanReq.ChannelType() != "direct-tcpip" {
				log.Infof("Rejecting non direct-tcpip channel request from %s@%v", conn.User(), remoteAddr)
				chanReq.Reject(ssh.Prohibited, "direct-tcpip channels only (-L)")
				continue
			}
			args := forwardChannelArgs{}
			err = ssh.Unmarshal(chanReq.ExtraData(), &args)
			if err != nil {
				log.Warningf("Failed to parse channel request data: %s", err)
				chanReq.Reject(ssh.Prohibited, "invalid request data")
				continue
			}
			if args.Port != tunnel.Link.Port {
				log.Warningf("Rejecting channel request on unexpected port %i, expected %i", args.Port, tunnel.Link.Port)
				chanReq.Reject(ssh.Prohibited, "bad port")
				continue
			}
			log.Infof("Forward request to port %s", args.Port)
			srcChan, srcReqs, err := chanReq.Accept()
			if err != nil {
				log.Error("Failed to accept direct-tcpip channel: %v", err)
				continue
			}
			log.Infof("Accepted direct-tcpip channel from %s@%v", conn.User(), remoteAddr)
			go ssh.DiscardRequests(srcReqs)

			dstConn := tunnel.Link.To.Conn
			if dstConn == nil {
				log.Warning("target not connected")
				continue
			}
			log.Infof("requesting forwarded-tcpip channel from %v to %v", tunnel.Link.From, tunnel.Link.To)
			dstChan, dstReqs, err := dstConn.OpenChannel("forwarded-tcpip", ssh.Marshal(&forwardChannelArgs{
				Addr:       "localhost",
				Port:       tunnel.Link.Port,
				OriginAddr: remoteAddr.IP.String(),
				OriginPort: uint32(remoteAddr.Port),
			}))
			if err != nil {
				log.Errorf("Failed to open dst channel: %s", err)
				continue
			}
			go ssh.DiscardRequests(dstReqs)
			go io.Copy(srcChan, dstChan)
			go io.Copy(dstChan, srcChan)
		}
	} else {
		go hub.handleRChannels(tunnel, conn, chanReqs)
		hub.handleRRequests(tunnel, conn, reqs)
	}
	log.Infof("Disconnecting user=%s", tunnel.User)
	return nil
}

func (hub *Hub) serializeLinks() []interface{} {
	links := make([]interface{}, len(hub.links))
	for i, link := range hub.links {
		links[i] = &struct {
			Port uint32       `json:"port"`
			From TunnelStatus `json:"from"`
			To   TunnelStatus `json:"to"`
		}{
			link.Port,
			link.From.Serialize(),
			link.To.Serialize(),
		}
	}
	return links
}
