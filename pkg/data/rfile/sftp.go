package rfile

import (
	"fmt"
	"io/ioutil"

	"net"
	"os"

	"github.com/pkg/sftp"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

const size = 32768

//SSHConfig contains SSH parameters
type SSHConfig struct {
	User string
	Key  string
}

//InitSFTP init sftp session
func InitSFTP(host string, sshUser string, key string) (*sftp.Client, error) {
	var auths []ssh.AuthMethod

	if IsFile(key) {
		pemBytes, err := ioutil.ReadFile(key)
		if err != nil {
			return nil, fmt.Errorf("Error on read KeyFile: Error: %s", err)
		}
		signer, err := ssh.ParsePrivateKey(pemBytes)
		if err != nil {
			return nil, fmt.Errorf("parse key failed: %s", err)
		}

		auths = append(auths, ssh.PublicKeys(signer))
	}

	// ssh agent support
	if aconn, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
		auths = append(auths, ssh.PublicKeysCallback(agent.NewClient(aconn).Signers))
	}

	config := &ssh.ClientConfig{
		User:            sshUser,
		Auth:            auths,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	sshhost := fmt.Sprintf("%s:22", host)
	conn, err := ssh.Dial("tcp", sshhost, config)
	if err != nil {
		return nil, fmt.Errorf("dial failed:%v", err)
	}

	c, err := sftp.NewClient(conn, sftp.MaxPacket(size))
	if err != nil {
		return nil, fmt.Errorf("unable to start sftp subsytem: %v", err)
	}
	return c, nil
}

//IsFile returns true if the file doesn't exist
func IsFile(file string) bool {
	stat, err := os.Stat(file)
	if err != nil {
		return false
	}
	if stat.Mode().IsRegular() {
		return true
	}

	return false
}
