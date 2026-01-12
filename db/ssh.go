package db

import (
	"fmt"
	"net"
	"os"

	"github.com/qyinm/lazyadmin/config"
	"golang.org/x/crypto/ssh"
)

type SSHTunnel struct {
	client   *ssh.Client
	listener net.Listener
	LocalAddr string
}

func NewSSHTunnel(cfg *config.SSHConfig, remoteHost string, remotePort int) (*SSHTunnel, error) {
	var authMethods []ssh.AuthMethod

	if cfg.PrivateKey != "" {
		key, err := os.ReadFile(cfg.PrivateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to read private key: %w", err)
		}

		var signer ssh.Signer
		if cfg.Password != "" {
			signer, err = ssh.ParsePrivateKeyWithPassphrase(key, []byte(cfg.Password))
		} else {
			signer, err = ssh.ParsePrivateKey(key)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}

	if cfg.Password != "" && cfg.PrivateKey == "" {
		authMethods = append(authMethods, ssh.Password(cfg.Password))
	}

	sshConfig := &ssh.ClientConfig{
		User:            cfg.User,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	sshAddr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	client, err := ssh.Dial("tcp", sshAddr, sshConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to dial SSH: %w", err)
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to create local listener: %w", err)
	}

	tunnel := &SSHTunnel{
		client:    client,
		listener:  listener,
		LocalAddr: listener.Addr().String(),
	}

	remoteAddr := fmt.Sprintf("%s:%d", remoteHost, remotePort)

	go func() {
		for {
			localConn, err := listener.Accept()
			if err != nil {
				return
			}

			remoteConn, err := client.Dial("tcp", remoteAddr)
			if err != nil {
				localConn.Close()
				continue
			}

			go tunnel.forward(localConn, remoteConn)
		}
	}()

	return tunnel, nil
}

func (t *SSHTunnel) forward(local, remote net.Conn) {
	defer local.Close()
	defer remote.Close()

	done := make(chan struct{}, 2)

	go func() {
		buf := make([]byte, 32*1024)
		for {
			n, err := remote.Read(buf)
			if n > 0 {
				local.Write(buf[:n])
			}
			if err != nil {
				break
			}
		}
		done <- struct{}{}
	}()

	go func() {
		buf := make([]byte, 32*1024)
		for {
			n, err := local.Read(buf)
			if n > 0 {
				remote.Write(buf[:n])
			}
			if err != nil {
				break
			}
		}
		done <- struct{}{}
	}()

	<-done
}

func (t *SSHTunnel) Close() error {
	if t.listener != nil {
		t.listener.Close()
	}
	if t.client != nil {
		return t.client.Close()
	}
	return nil
}
