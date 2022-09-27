package main

// Forward from local port 9000 to remote port 9999

import (
	"flag"
	"io"
	"log"
	"net"

	"golang.org/x/crypto/ssh"
)

var (
	username         = "ender"
	serverAddrString = "209.181.69.247:33122"
	localAddrString  = "localhost:8545"
	remoteAddrString = "localhost:8545"
)

type Connection struct {
	username         string
	serverAddrString string
	localAddrString  string
	remoteAddrString string
}

func forward(localConn net.Conn, config *ssh.ClientConfig) {
	// Setup sshClientConn (type *ssh.ClientConn)
	sshClientConn, err := ssh.Dial("tcp", serverAddrString, config)
	if err != nil {
		log.Fatalf("ssh.Dial failed: %s", err)
	}

	// Setup sshConn (type net.Conn)
	sshConn, err := sshClientConn.Dial("tcp", remoteAddrString)

	if err != nil {
		log.Println("Port forward successful")
	}

	// Copy localConn.Reader to sshConn.Writer
	go func() {
		_, err = io.Copy(sshConn, localConn)
		if err != nil {
			log.Fatalf("io.Copy failed: %v", err)
		}
	}()

	// Copy sshConn.Reader to localConn.Writer
	go func() {
		_, err = io.Copy(localConn, sshConn)
		if err != nil {
			log.Fatalf("io.Copy failed: %v", err)
		}
	}()
}

var hostFlag string
var portsFlag string

func main() {
	log.Println("Starting ssh tunnel v0.0.1")

	flag.StringVar(&hostFlag, "host", "localhost", "The remove server address.")
	flag.StringVar(&portsFlag, "ports", "8545", "The ports to forward from the remote, in a comma separated list.")

	// get flags

	key, err := parsePrivateKey(privateKeyPath())
	if err != nil {
		log.Fatalf("parsePrivateKey failed: %v", err)
	}
	log.Println("Loaded private key at " + privateKeyPath())

	// Setup SSH config (type *ssh.ClientConfig)
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(key),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// Setup localListener (type net.Listener)
	localListener, err := net.Listen("tcp", localAddrString)
	if err != nil {
		log.Fatalf("net.Listen failed: %v", err)
	}

	log.Println("Port forward successful")

	for {
		// Setup localConn (type net.Conn)
		localConn, err := localListener.Accept()
		if err != nil {
			log.Fatalf("listen.Accept failed: %v", err)
		}
		go forward(localConn, config)
	}
}
