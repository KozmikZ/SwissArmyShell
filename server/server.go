package server

import (
	"bytes"
	"strconv"
	"strings"

	"golang.org/x/crypto/ssh"
)

func ConnectSSH(username string, password string, target string) (ServerSession, error) {
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	client, err := ssh.Dial("tcp", target, config)
	if err != nil {
		println("Fatal error during ssh connection")
		return ServerSession{}, err
	}
	var wd string
	var b bytes.Buffer
	session, err := client.NewSession()
	if err != nil {
		println("Failed to estabilish session")
	}
	defer session.Close()
	session.Stdout = &b
	erro := session.Run("pwd")
	if erro != nil {
		println(erro)
	}
	wd = b.String()[:len(b.String())-1]
	return ServerSession{ssh: client, config: config, Wd: wd}, err
}

type ServerSession struct {
	config *ssh.ClientConfig
	ssh    *ssh.Client
	Wd     string
}

type File struct {
	Name     string
	Mode     string
	Modified string
	Size     int64
	IsDir    bool
}

func (s *ServerSession) ListFiles() []File {
	rawOut, _ := s.ExecuteRaw("ls -la")
	rawStrOut := string(rawOut)
	rawEntries := strings.Split(rawStrOut, "\n")
	numEntries := len(rawEntries) - 1
	files := []File{}
	for i := 1; i < numEntries; i++ {
		newRawEntry := rawEntries[i]
		fields := strings.Split(newRawEntry, " ")
		IsDir := string(fields[0][0]) == "d"
		var dateMod string
		var Size int64
		if IsDir {
			dateMod = fields[5] + " " + fields[7] + " " + fields[8]
			Size = 4096
		} else {
			dateMod = fields[6] + " " + fields[8] + " " + fields[9]
			tmp, _ := strconv.Atoi(fields[5])
			Size = int64(tmp)
		}

		files = append(files, File{Name: fields[len(fields)-1],
			Mode:     fields[0],
			Modified: dateMod,
			Size:     Size,
			IsDir:    IsDir})
	}
	return files
}
func (s *ServerSession) GetWD() string {
	rawOut, _ := s.ExecuteRaw("pwd")
	s.Wd = rawOut[:len(rawOut)-1]
	return s.Wd
}

func (s *ServerSession) ChangeWD(WD string) {
	str, _ := s.ExecuteRaw("cd " + WD + " && pwd")
	s.Wd = str
}

func (s *ServerSession) ReadFileInput(name string) (string, error) {
	rawOut, err := s.ExecuteRaw("cat " + name)
	return string(rawOut), err
}

func (s *ServerSession) ReWriteFile(name string, content string) {
	s.ExecuteRaw("echo " + content + "> " + name)
}

func (s *ServerSession) ExecuteRaw(command string) (string, error) {
	var b bytes.Buffer
	session, err := s.ssh.NewSession()
	if err != nil {
		println("Failed to estabilish session")
	}
	defer session.Close()
	finalCmd := "cd " + s.Wd + " && " + command
	finalCmd = strings.TrimRight(finalCmd, "\n")
	// Attach buffer to session's stdout and stderr
	session.Stdout = &b
	session.Stderr = &b
	println(finalCmd)
	err = session.Run(finalCmd)
	if err != nil {
		println("Error ocurred")
		println(err)
		return "", err
	} else {
		println(b.String())
	}
	return b.String(), err
}
