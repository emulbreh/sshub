package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/voltaro/sshub/libsshub"
	"golang.org/x/crypto/ssh"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
	"io/ioutil"
	"net/http"
	"os"
)

var (
	privateKeyFile = kingpin.Flag("key-file", "Private key of the server").Short('k').Default("/etc/sshub/key").String()
	address        = kingpin.Flag("address", "Listen address of the SSH server").Default(":4022").TCP()
	httpAddress    = kingpin.Flag("http", "Listen address of the HTTP api").Default(":4080").TCP()
)

func main() {
	kingpin.Parse()
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})

	log.Infof("Loading private key from %s", *privateKeyFile)
	privateBytes, err := ioutil.ReadFile(*privateKeyFile)
	if err != nil {
		log.Errorf("Failed to load private: %v", *privateKeyFile, err)
		os.Exit(1)
	}
	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		panic("Failed to parse private key")
	}

	log.Infof("Listening for ssh connections on %v", *address)
	hub := libsshub.NewHub(private)
	go hub.Listen((*address).String())

	log.Infof("Listening for http connections on %v", *httpAddress)
	libsshub.InstallHttpHandlers(hub)
	err = http.ListenAndServe((*httpAddress).String(), nil)
	if err != nil {
		log.Errorf("Failed to serve http api: %s", err)
	}

}
