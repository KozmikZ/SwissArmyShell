package server_ssh

import (
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
	session, err := client.NewSession()
	if err != nil {
		println("Failed to estabilish session")
	}
	return ServerSession{ssh: session}, err
}

type ServerSession struct {
	ssh *ssh.Session
}

type File struct {
	name     string
	mode     string
	modified string
	size     int
	isDir    bool
}

func (s ServerSession) ListFiles() []File {
	rawOut, _ := s.ssh.CombinedOutput("ls -la")
	rawStrOut := string(rawOut)
	rawEntries := strings.Split(rawStrOut, "\n")
	numEntries := len(rawEntries) - 1
	files := []File{}
	for i := 1; i < numEntries; i++ {
		newRawEntry := rawEntries[i]
		fields := strings.Split(newRawEntry, " ")
		isDir := string(fields[0][0]) == "d"
		var dateMod string
		var size int
		if isDir {
			dateMod = fields[5] + " " + fields[7] + " " + fields[8]
			size = 4096
		} else {
			dateMod = fields[6] + " " + fields[8] + " " + fields[9]
			size, _ = strconv.Atoi(fields[5])
		}

		files = append(files, File{name: fields[len(fields)-1],
			mode:     fields[0],
			modified: dateMod,
			size:     size,
			isDir:    isDir})
	}
	return files
}
func (s ServerSession) GetWD() string {
	rawOut, _ := s.ssh.CombinedOutput("pwd")
	return string(rawOut)
}

func (s ServerSession) ChangeWD(WD string) {
	s.ssh.CombinedOutput("cd " + WD)
}

func (s ServerSession) readFileInput(name string) string {
	rawOut, _ := s.ssh.CombinedOutput("cat " + name)
	return string(rawOut)
}

func (s ServerSession) reWriteFile(name string, content string) {
	s.ssh.CombinedOutput("echo " + content + "> " + name)
}

func (s ServerSession) executeRaw(command string) ([]byte, error) {
	out, err := s.ssh.CombinedOutput(command)
	return out, err
}

func main() {
	session, _ := ConnectSSH("test", "test", "localhost:22")
	for _, file := range session.ListFiles() {
		println(file.name)
		println(file.isDir)
		println(file.mode)
		println(file.modified)
		println(file.size)
	}

}
