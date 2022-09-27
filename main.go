package main

// Forward from local port 9000 to remote port 9999

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os/user"
	"strconv"

	"golang.org/x/crypto/ssh"
)

// var (
// 	username         = "ender"
// 	serverAddrString = "209.181.69.247:33122"
// )

var hostFlag string
var portsFlag int
var usernameFlag string
var keyPathFlag string

func main() {
	currentUser, err := user.Current()
	if err != nil {
		log.Fatalf("Faile to get current user: %v", err)
	}

	flag.StringVar(&hostFlag, "h", "", "The remove host address. e.g. 10.0.0.11:22")
	flag.IntVar(&portsFlag, "p", -777, "The ports to forward from the remote.")
	flag.StringVar(&usernameFlag, "u", currentUser.Username, "Username to log in with.")
	flag.StringVar(&keyPathFlag, "k", privateKeyPath(), "Path to the private key to use.")

	flag.Parse()

	if hostFlag == "" {
		log.Fatalf("Host is required.")
	}
	if portsFlag == -777 {
		log.Fatalf("Port to forward is required.")
	}
	if (portsFlag < 1) || (portsFlag > 65535) {
		log.Fatalf("Port to forward is invalid.")
	}

	log.Println("Starting ssh tunnel v0.0.1")
	fmt.Println("Config: ", hostFlag, portsFlag, usernameFlag, keyPathFlag)

	key, err := parsePrivateKey(keyPathFlag)
	if err != nil {
		log.Fatalf("Parsing private key failed: %v", err)
	}
	log.Println("Loaded private key at " + privateKeyPath())

	// Setup SSH config (type *ssh.ClientConfig)
	config := &ssh.ClientConfig{
		User: usernameFlag,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(key),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// Setup localListener (type net.Listener)
	localListener, err := net.Listen("tcp", "localhost:"+strconv.Itoa(portsFlag))
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

func forward(localConn net.Conn, config *ssh.ClientConfig) {
	// Setup sshClientConn (type *ssh.ClientConn)
	sshClientConn, err := ssh.Dial("tcp", "localhost:"+strconv.Itoa(portsFlag), config)
	if err != nil {
		log.Fatalf("ssh.Dial failed: %s", err)
	}

	// Setup sshConn (type net.Conn)
	sshConn, err := sshClientConn.Dial("tcp", strconv.Itoa(portsFlag))

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
