package util

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// ExecReturnStdOut executes a command and returns stdout of the command
func ExecReturnStdOut(cmdToExecute []string) (string, error) {

	var cmd *exec.Cmd
	if length := len(cmdToExecute); length == 1 {
		cmd = exec.Command(cmdToExecute[0])
	} else {
		cmd = exec.Command(cmdToExecute[0], cmdToExecute[1:]...)
	}

	var stdErr bytes.Buffer
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stdErr

	err := cmd.Run()
	if err != nil {
		return "", errors.Join(err, fmt.Errorf("stderr: %s", stdErr.String()))
	}

	return strings.TrimSuffix(out.String(), "\n"), nil
}
