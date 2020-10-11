package auth

import (
	"jjz.io/xscp/client"
	"time"

	"golang.org/x/crypto/ssh"
)

// A struct containing all the configuration options
// used by an client client.
type ClientConfigurer struct {
	host         string
	clientConfig *ssh.ClientConfig
	session      *ssh.Session
	timeout      time.Duration
	remoteBinary string
}

// Creates a new client configurer.
// It takes the required parameters: the host and the ssh.ClientConfig and
// returns a configurer populated with the default values for the optional
// parameters.
//
// These optional parameters can be set by using the methods provided on the
// ClientConfigurer struct.
func NewConfigurer(host string, config *ssh.ClientConfig) *ClientConfigurer {
	return &ClientConfigurer{
		host:         host,
		clientConfig: config,
		timeout:      time.Minute,
		remoteBinary: "scp", //todo
	}
}

// Sets the path of the location of the remote client binary
// Defaults to: /usr/bin/client
func (c *ClientConfigurer) RemoteBinary(path string) *ClientConfigurer {
	c.remoteBinary = path
	return c
}

// Alters the host of the client connects to
func (c *ClientConfigurer) Host(host string) *ClientConfigurer {
	c.host = host
	return c
}

// Changes the connection timeout.
// Defaults to one minute
func (c *ClientConfigurer) Timeout(timeout time.Duration) *ClientConfigurer {
	c.timeout = timeout
	return c
}

// Alters the ssh.ClientConfig
func (c *ClientConfigurer) ClientConfig(config *ssh.ClientConfig) *ClientConfigurer {
	c.clientConfig = config
	return c
}

// Alters the ssh.Session
func (c *ClientConfigurer) Session(session *ssh.Session) *ClientConfigurer {
	c.session = session
	return c
}

// Builds a client with the configuration stored within the ClientConfigurer
func (c *ClientConfigurer) Create() client.Client {
	return client.Client{
		Host:         c.host,
		ClientConfig: c.clientConfig,
		Timeout:      c.timeout,
		RemoteBinary: c.remoteBinary,
		Session:      c.session,
	}
}

// Returns a new client.Client with provided host and ssh.clientConfig
// It has a default timeout of one minute.
func NewClient(host string, config *ssh.ClientConfig) client.Client {
	return NewConfigurer(host, config).Create()
}

// Returns a new client.Client with provides host, ssh.ClientConfig and timeout
func NewClientWithTimeout(host string, config *ssh.ClientConfig, timeout time.Duration) client.Client {
	return NewConfigurer(host, config).Timeout(timeout).Create()
}

// Returns a new client.Client using an already existing established SSH connection
func NewClientBySSH(ssh *ssh.Client) (client.Client, error) {
	session, err := ssh.NewSession()
	if err != nil {
		return client.Client{}, err
	}
	return NewConfigurer("", nil).Session(session).Create(), nil
}

/// Same as NewClientWithTimeout but uses an existing SSH client
func NewClientBySSHWithTimeout(ssh *ssh.Client, timeout time.Duration) (client.Client, error) {
	session, err := ssh.NewSession()
	if err != nil {
		return client.Client{}, err
	}
	return NewConfigurer("", nil).Session(session).Timeout(timeout).Create(), nil
}
