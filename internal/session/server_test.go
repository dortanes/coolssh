package session

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"io"
	"net"
	"strconv"
	"testing"

	"golang.org/x/crypto/ssh"
)

// testServer is an SSH server on the loopback interface whose shell echoes what it is sent and
// exits on "exit". It forwards direct-tcpip channels, so it can be a jump host.
type testServer struct {
	listener net.Listener
	hostKey  ssh.Signer
	config   *ssh.ServerConfig
	// silent leaves global requests, keepalives among them, unanswered.
	silent bool
}

func newSigner(t *testing.T) ssh.Signer {
	t.Helper()
	_, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	signer, err := ssh.NewSignerFromKey(private)
	if err != nil {
		t.Fatal(err)
	}
	return signer
}

func newServer(t *testing.T, configure func(*ssh.ServerConfig)) *testServer {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { listener.Close() })
	server := &testServer{listener: listener, hostKey: newSigner(t), config: &ssh.ServerConfig{}}
	configure(server.config)
	server.config.AddHostKey(server.hostKey)
	go server.accept()
	return server
}

func (s *testServer) port() int {
	return s.listener.Addr().(*net.TCPAddr).Port
}

// address is how known_hosts names the server.
func (s *testServer) address() string {
	return net.JoinHostPort("127.0.0.1", strconv.Itoa(s.port()))
}

func (s *testServer) accept() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return
		}
		go s.serve(conn)
	}
}

func (s *testServer) serve(conn net.Conn) {
	_, channels, requests, err := ssh.NewServerConn(conn, s.config)
	if err != nil {
		conn.Close()
		return
	}
	go func() {
		for request := range requests {
			if !s.silent {
				_ = request.Reply(true, nil)
			}
		}
	}()
	for channel := range channels {
		switch channel.ChannelType() {
		case "session":
			accepted, requests, err := channel.Accept()
			if err == nil {
				go shell(accepted, requests)
			}
		case "direct-tcpip":
			go forward(channel)
		default:
			_ = channel.Reject(ssh.UnknownChannelType, "")
		}
	}
}

func shell(channel ssh.Channel, requests <-chan *ssh.Request) {
	for request := range requests {
		_ = request.Reply(true, nil)
		switch request.Type {
		case "shell":
			go echo(channel)
		case "exec":
			go run(channel, request.Payload)
		}
	}
}

// run answers an exec request: the command "false" exits with status 1, any other with 0.
func run(channel ssh.Channel, payload []byte) {
	defer channel.Close()
	var command struct{ Command string }
	_ = ssh.Unmarshal(payload, &command)
	status := uint32(0)
	if command.Command == "false" {
		status = 1
	}
	_, _ = io.WriteString(channel, "ran\r\n")
	_, _ = channel.SendRequest("exit-status", false, ssh.Marshal(struct{ Status uint32 }{status}))
}

func echo(channel ssh.Channel) {
	defer channel.Close()
	_, _ = io.WriteString(channel, "ready\r\n")
	buffer := make([]byte, 1024)
	for {
		n, err := channel.Read(buffer)
		if err != nil {
			return
		}
		if bytes.Contains(buffer[:n], []byte("exit")) {
			_, _ = channel.SendRequest("exit-status", false, ssh.Marshal(struct{ Status uint32 }{0}))
			return
		}
		_, _ = channel.Write(buffer[:n])
	}
}

func forward(channel ssh.NewChannel) {
	var target struct {
		Host       string
		Port       uint32
		OriginHost string
		OriginPort uint32
	}
	if err := ssh.Unmarshal(channel.ExtraData(), &target); err != nil {
		_ = channel.Reject(ssh.ConnectionFailed, err.Error())
		return
	}
	conn, err := net.Dial("tcp", net.JoinHostPort(target.Host, strconv.Itoa(int(target.Port))))
	if err != nil {
		_ = channel.Reject(ssh.ConnectionFailed, err.Error())
		return
	}
	accepted, requests, err := channel.Accept()
	if err != nil {
		conn.Close()
		return
	}
	go ssh.DiscardRequests(requests)
	go func() {
		_, _ = io.Copy(accepted, conn)
		accepted.Close()
	}()
	_, _ = io.Copy(conn, accepted)
	conn.Close()
}

// acceptPassword is a server configuration taking only password.
func acceptPassword(password string) func(*ssh.ServerConfig) {
	return func(config *ssh.ServerConfig) {
		config.PasswordCallback = func(_ ssh.ConnMetadata, given []byte) (*ssh.Permissions, error) {
			if string(given) == password {
				return nil, nil
			}
			return nil, errRejected
		}
	}
}

// acceptKey is a server configuration taking only key.
func acceptKey(key ssh.PublicKey) func(*ssh.ServerConfig) {
	return func(config *ssh.ServerConfig) {
		config.PublicKeyCallback = func(_ ssh.ConnMetadata, given ssh.PublicKey) (*ssh.Permissions, error) {
			if bytes.Equal(given.Marshal(), key.Marshal()) {
				return nil, nil
			}
			return nil, errRejected
		}
	}
}

var errRejected = io.ErrUnexpectedEOF
