package ssh

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"time"

	"golang.org/x/crypto/ssh"
)

// Config holds SSH connection configuration
type Config struct {
	Host        string
	Port        int
	Username    string
	PrivateKey  string
	BastionHost string
	BastionPort int
	BastionUser string
	BastionKey  string
	Timeout     time.Duration
	KeepAlive   time.Duration
}

// Client represents an SSH client
type Client struct {
	config *Config
	client *ssh.Client
}

// NewClient creates a new SSH client
func NewClient(config *Config) (*Client, error) {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.KeepAlive == 0 {
		config.KeepAlive = 10 * time.Second
	}

	return &Client{
		config: config,
	}, nil
}

// Connect establishes SSH connection
func (c *Client) Connect() error {
	signer, err := ssh.ParsePrivateKey([]byte(c.config.PrivateKey))
	if err != nil {
		return fmt.Errorf("failed to parse private key: %w", err)
	}

	sshConfig := &ssh.ClientConfig{
		User: c.config.Username,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO: Implement proper host key verification
		Timeout:         c.config.Timeout,
	}

	// Direct connection or through bastion
	if c.config.BastionHost != "" {
		return c.connectThroughBastion(sshConfig)
	}

	addr := fmt.Sprintf("%s:%d", c.config.Host, c.config.Port)
	client, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", addr, err)
	}

	c.client = client
	return nil
}

// connectThroughBastion connects through a bastion host
func (c *Client) connectThroughBastion(targetConfig *ssh.ClientConfig) error {
	// Parse bastion key
	bastionSigner, err := ssh.ParsePrivateKey([]byte(c.config.BastionKey))
	if err != nil {
		return fmt.Errorf("failed to parse bastion key: %w", err)
	}

	bastionConfig := &ssh.ClientConfig{
		User: c.config.BastionUser,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(bastionSigner),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         c.config.Timeout,
	}

	// Connect to bastion
	bastionAddr := fmt.Sprintf("%s:%d", c.config.BastionHost, c.config.BastionPort)
	bastionClient, err := ssh.Dial("tcp", bastionAddr, bastionConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to bastion %s: %w", bastionAddr, err)
	}

	// Connect to target through bastion
	targetAddr := fmt.Sprintf("%s:%d", c.config.Host, c.config.Port)
	conn, err := bastionClient.Dial("tcp", targetAddr)
	if err != nil {
		bastionClient.Close()
		return fmt.Errorf("failed to connect to target through bastion: %w", err)
	}

	ncc, chans, reqs, err := ssh.NewClientConn(conn, targetAddr, targetConfig)
	if err != nil {
		conn.Close()
		bastionClient.Close()
		return fmt.Errorf("failed to establish SSH connection: %w", err)
	}

	c.client = ssh.NewClient(ncc, chans, reqs)
	return nil
}

// ExecuteCommand executes a command on the remote host
func (c *Client) ExecuteCommand(ctx context.Context, command string) (stdout, stderr string, exitCode int, err error) {
	if c.client == nil {
		return "", "", -1, fmt.Errorf("not connected")
	}

	session, err := c.client.NewSession()
	if err != nil {
		return "", "", -1, fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	var stdoutBuf, stderrBuf bytes.Buffer
	session.Stdout = &stdoutBuf
	session.Stderr = &stderrBuf

	// Run command with context
	errChan := make(chan error, 1)
	go func() {
		errChan <- session.Run(command)
	}()

	select {
	case <-ctx.Done():
		_ = session.Signal(ssh.SIGTERM) // Best effort signal
		return stdoutBuf.String(), stderrBuf.String(), -1, ctx.Err()
	case err := <-errChan:
		if err != nil {
			if exitErr, ok := err.(*ssh.ExitError); ok {
				return stdoutBuf.String(), stderrBuf.String(), exitErr.ExitStatus(), nil
			}
			return stdoutBuf.String(), stderrBuf.String(), -1, err
		}
		return stdoutBuf.String(), stderrBuf.String(), 0, nil
	}
}

// CreateTunnel creates an SSH tunnel
func (c *Client) CreateTunnel(localAddr, remoteAddr string) (*Tunnel, error) {
	if c.client == nil {
		return nil, fmt.Errorf("not connected")
	}

	listener, err := net.Listen("tcp", localAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to create local listener: %w", err)
	}

	tunnel := &Tunnel{
		client:     c,
		listener:   listener,
		remoteAddr: remoteAddr,
		closeCh:    make(chan struct{}),
	}

	go tunnel.serve()

	return tunnel, nil
}

// Close closes the SSH connection
func (c *Client) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

// Tunnel represents an SSH tunnel
type Tunnel struct {
	client     *Client
	listener   net.Listener
	remoteAddr string
	closeCh    chan struct{}
}

// serve handles tunnel connections
func (t *Tunnel) serve() {
	for {
		select {
		case <-t.closeCh:
			return
		default:
			localConn, err := t.listener.Accept()
			if err != nil {
				select {
				case <-t.closeCh:
					return
				default:
					continue
				}
			}

			go t.handleConnection(localConn)
		}
	}
}

// handleConnection handles a single tunnel connection
func (t *Tunnel) handleConnection(localConn net.Conn) {
	defer localConn.Close()

	remoteConn, err := t.client.client.Dial("tcp", t.remoteAddr)
	if err != nil {
		return
	}
	defer remoteConn.Close()

	// Bidirectional copy
	done := make(chan struct{}, 2)

	go func() {
		_, _ = io.Copy(remoteConn, localConn)
		done <- struct{}{}
	}()

	go func() {
		_, _ = io.Copy(localConn, remoteConn)
		done <- struct{}{}
	}()

	<-done
}

// Close closes the tunnel
func (t *Tunnel) Close() error {
	close(t.closeCh)
	return t.listener.Close()
}

// LocalAddr returns the local address of the tunnel
func (t *Tunnel) LocalAddr() string {
	return t.listener.Addr().String()
}
