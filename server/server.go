package server

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path"
	"strings"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// ConnectSSH establishes an SSH connection and returns a ServerSession object.
func ConnectSSH(username, password, target, hostKeyPath string) (ServerSession, error) {
	var hostKeyCallback ssh.HostKeyCallback
	var err error
	if hostKeyPath == "" {
		hostKeyCallback = ssh.InsecureIgnoreHostKey()
	} else {
		hostKeyCallback, err = knownhosts.New(hostKeyPath)
	}
	if err != nil {
		return ServerSession{}, err // improved host key verification
	}

	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: hostKeyCallback, // using hostKeyCallback instead of ssh.InsecureIgnoreHostKey()
	}

	client, err := ssh.Dial("tcp", target, config)
	if err != nil {
		return ServerSession{}, err
	}

	session, err := client.NewSession()
	if err != nil {
		return ServerSession{}, err
	}
	defer session.Close()

	var b bytes.Buffer
	session.Stdout = &b
	if err := session.Run("pwd"); err != nil {
		return ServerSession{}, err
	}

	wd := strings.TrimSpace(b.String())
	return ServerSession{ssh: client, config: config, Wd: wd}, nil
}

func (s *ServerSession) ConnectSFTP() error {
	// Create an SFTP client
	sftpClient, err := sftp.NewClient(s.ssh)
	if err != nil {
		return err
	}
	s.sftp = sftpClient
	return nil
}

type ServerSession struct {
	config *ssh.ClientConfig
	ssh    *ssh.Client
	sftp   *sftp.Client
	Wd     string
}

type File struct {
	Name     string
	Mode     string
	Modified string
	Size     int64
	IsDir    bool
}

func (s *ServerSession) ListFiles() ([]File, error) {
	files := []File{}

	// List the files in the current directory
	fileInfos, err := s.sftp.ReadDir(s.Wd)
	if err != nil {
		return nil, err
	}

	for _, fileInfo := range fileInfos {
		file := File{
			Name:     fileInfo.Name(),
			Mode:     fileInfo.Mode().String(),
			Modified: fileInfo.ModTime().String(),
			Size:     fileInfo.Size(),
			IsDir:    fileInfo.IsDir(),
		}
		files = append(files, file)
	}

	return files, nil
}

// GetWD returns the current working directory.
func (s *ServerSession) GetWD() string {
	return s.Wd
}

func (s *ServerSession) ChangeWD(newWD string) error {
	// Check if the new working directory exists and is a directory
	fileInfo, err := s.sftp.Stat(path.Join(s.Wd, newWD))
	if err != nil {
		return err // The directory does not exist or some other error
	}
	if !fileInfo.IsDir() {
		return fmt.Errorf("not a directory: %s", newWD)
	}

	// Update the current working directory
	s.Wd = path.Join(s.Wd, newWD)
	return nil
}

func (s *ServerSession) ReadFileInput(filePath string) (string, error) {
	// Ensure the file path is absolute or properly relative to `Wd`
	fullPath := path.Join(s.Wd, filePath)

	file, err := s.sftp.Open(fullPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	contents, err := ioutil.ReadAll(file)
	if err != nil {
		return "", err
	}

	return string(contents), nil
}

func (s *ServerSession) ReWriteFile(filePath string, content string) error {
	// Ensure the file path is absolute or properly relative to `Wd`
	fullPath := path.Join(s.Wd, filePath)

	file, err := s.sftp.Create(fullPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write([]byte(content))
	return err
}

func (s *ServerSession) RemoveFileTarget(target string) error {
	fullPath := path.Join(s.Wd, target)
	isDirectory := false
	_, err := s.sftp.ReadDir(fullPath)
	if err != nil {
		isDirectory = true
	}
	var err2 error
	if isDirectory {
		err2 = s.sftp.RemoveDirectory(fullPath)
	} else {
		err2 = s.sftp.Remove(fullPath)
	}
	return err2
}

// ExecuteRaw executes a command in the server's shell and returns its string output
func (s *ServerSession) ExecuteRaw(command string) (string, error) {
	session, err := s.ssh.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	var b bytes.Buffer
	session.Stdout = &b
	session.Stderr = &b

	if err := session.Run(command); err != nil {
		return "", err
	}

	return b.String(), err
}
