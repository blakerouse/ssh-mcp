package ssh

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/knownhosts"
)

// ErrNotConnected returned when the client is not connected.
var ErrNotConnected = errors.New("not connected")

// OSInfo provides the OS information.
type OSInfo struct {
	OSRelease string `yaml:"os_release" json:"os_release" jsonschema_description:"The output of /etc/os-release"`
	Uname     string `yaml:"uname" json:"uname" jsonschema_description:"The output of the uname command"`
}

// ClientInfo stores the generate client information.
type ClientInfo struct {
	Name  string `yaml:"name" json:"name" jsonschema_description:"The name of the client"`
	Group string `yaml:"group" json:"group" jsonschema_description:"The group the client belongs to"`
	Host  string `yaml:"host" json:"host" jsonschema_description:"The host of the client"`
	Port  string `yaml:"port" json:"port" jsonschema_description:"The port of the client"`
	User  string `yaml:"user" json:"user" jsonschema_description:"The user of the client (optional, defaults to current user)"`
	Pass  string `yaml:"pass,omitempty" json:"pass,omitempty" jsonschema_description:"The password of the client (optional, will use SSH agent if not provided)"`

	OS OSInfo `yaml:"os" json:"os" jsonschema_description:"The operating system information"`
}

// NewClientInfo returns client information from the connection string.
func NewClientInfo(name string, connStr string) (*ClientInfo, error) {
	sshURL, err := url.Parse(connStr)
	if err != nil {
		return nil, fmt.Errorf("invalid SSH connection string: %w", err)
	}
	if sshURL.Scheme != "ssh" {
		return nil, errors.New("invalid SSH connection string: not ssh scheme")
	}

	// Username is optional - will default to current user if not provided
	user := ""
	pass := ""
	if sshURL.User != nil {
		user = sshURL.User.Username()
		pass, _ = sshURL.User.Password()
	}

	host := sshURL.Hostname()
	if host == "" {
		return nil, errors.New("invalid SSH connection string: missing host")
	}

	port := sshURL.Port()
	if port == "" {
		port = "22" // default SSH port
	}
	if name == "" {
		name = host // default name to host (if not provided)
	}

	return &ClientInfo{
		Name: name,
		Host: host,
		Port: port,
		User: user,
		Pass: pass,
	}, nil
}

// Client is an SSH client.
type Client struct {
	info *ClientInfo

	client *ssh.Client
}

// NewClient creates the client with the hostPort and configuration.
func NewClient(info *ClientInfo) *Client {
	return &Client{
		info: info,
	}
}

// Connect connects to the SSH server.
func (c *Client) Connect() error {
	var err error
	host := fmt.Sprintf("%s:%s", c.info.Host, c.info.Port)

	// Use current user if not specified
	user := c.info.User
	if user == "" {
		user = os.Getenv("USER")
		if user == "" {
			user = os.Getenv("USERNAME") // Windows fallback
		}
	}

	// Build authentication methods
	authMethods := buildAuthMethods(c.info.Pass)

	// If no auth methods available, return error
	if len(authMethods) == 0 {
		return errors.New("no authentication method available: provide password, ensure SSH_AUTH_SOCK is set, or add SSH keys to ~/.ssh/")
	}

	// Get host key callback for secure host verification
	hostKeyCallback, err := getHostKeyCallback()
	if err != nil {
		return fmt.Errorf("failed to get host key callback: %w", err)
	}

	cfg := &ssh.ClientConfig{
		User:            user,
		Auth:            authMethods,
		HostKeyCallback: hostKeyCallback,
	}
	c.client, err = ssh.Dial("tcp", host, cfg)
	if err != nil {
		return fmt.Errorf("failed to connect to SSH server: %w", err)
	}
	return nil
}

// Close closes the connection to the SSH server.
func (c *Client) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

// Exec runs a command on the remote SSH server.
func (c *Client) Exec(cmd string) ([]byte, error) {
	session, err := c.client.NewSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()

	output, err := session.CombinedOutput(cmd)
	if err != nil {
		return nil, err
	}
	return output, nil
}

// loadPrivateKey loads a private key from a file
func loadPrivateKey(path string) (ssh.Signer, error) {
	key, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, err
	}

	return signer, nil
}

// buildAuthMethods builds a list of SSH authentication methods based on available credentials
func buildAuthMethods(password string) []ssh.AuthMethod {
	authMethods := []ssh.AuthMethod{}

	// If password is provided, use password authentication first
	if password != "" {
		authMethods = append(authMethods, ssh.Password(password))
	}

	// Try to use SSH agent
	sshAuthSock := os.Getenv("SSH_AUTH_SOCK")
	if sshAuthSock != "" {
		if agentConn, err := net.Dial("unix", sshAuthSock); err == nil {
			authMethods = append(authMethods, ssh.PublicKeysCallback(agent.NewClient(agentConn).Signers))
		}
	}

	// Try to load SSH keys from standard locations
	homeDir, err := os.UserHomeDir()
	if err == nil {
		keyPaths := []string{
			filepath.Join(homeDir, ".ssh", "id_rsa"),
			filepath.Join(homeDir, ".ssh", "id_ed25519"),
			filepath.Join(homeDir, ".ssh", "id_ecdsa"),
			filepath.Join(homeDir, ".ssh", "id_dsa"),
		}

		var signers []ssh.Signer
		for _, keyPath := range keyPaths {
			if signer, err := loadPrivateKey(keyPath); err == nil {
				signers = append(signers, signer)
			}
		}

		if len(signers) > 0 {
			authMethods = append(authMethods, ssh.PublicKeys(signers...))
		}
	}

	return authMethods
}

// getHostKeyCallback returns a HostKeyCallback that uses the known_hosts file
// It will automatically add new hosts to the known_hosts file
func getHostKeyCallback() (ssh.HostKeyCallback, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	knownHostsPath := filepath.Join(homeDir, ".ssh", "known_hosts")

	// Check if known_hosts file exists
	if _, err := os.Stat(knownHostsPath); os.IsNotExist(err) {
		// Create the .ssh directory if it doesn't exist
		sshDir := filepath.Join(homeDir, ".ssh")
		if err := os.MkdirAll(sshDir, 0700); err != nil {
			return nil, fmt.Errorf("failed to create .ssh directory: %w", err)
		}

		// Create an empty known_hosts file
		if _, err := os.Create(knownHostsPath); err != nil {
			return nil, fmt.Errorf("failed to create known_hosts file: %w", err)
		}
	}

	// Use the known_hosts file for host key verification
	hostKeyCallback, err := knownhosts.New(knownHostsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load known_hosts file: %w", err)
	}

	// Wrap the callback to automatically add new hosts
	return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		err := hostKeyCallback(hostname, remote, key)
		if err != nil {
			// Check if this is a "host key not found" error
			var keyErr *knownhosts.KeyError
			if errors.As(err, &keyErr) && len(keyErr.Want) == 0 {
				// Host not in known_hosts, add it
				f, fileErr := os.OpenFile(knownHostsPath, os.O_APPEND|os.O_WRONLY, 0600)
				if fileErr != nil {
					return fmt.Errorf("failed to open known_hosts for writing: %w", fileErr)
				}
				defer f.Close()

				// Format: hostname ssh-rsa AAAAB3N...
				line := knownhosts.Line([]string{hostname}, key)
				if _, writeErr := f.WriteString(line + "\n"); writeErr != nil {
					return fmt.Errorf("failed to write to known_hosts: %w", writeErr)
				}

				// Host was added, so accept this connection
				return nil
			}
			// Some other error (key mismatch, etc.)
			return err
		}
		// Host key matched
		return nil
	}, nil
}
