package shell

import (
	"fmt"
	"os"
	"os/exec"
)

func Init() {
	fmt.Println("Initialize subprocess")
}

func RunInShell(cmd string, args string) *exec.Cmd {
	return exec.Command(cmd, args)
}

func ListFiles() ([]os.DirEntry, error) {
	cwd, _ := os.Getwd()
	files, err := os.ReadDir(cwd)
	if err != nil {
		return []os.DirEntry{}, err
	} else {
		return files, nil
	}
}
