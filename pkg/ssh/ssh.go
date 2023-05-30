package ssh

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
)

// CopyToRemote copies a file or directory to a remote machine
func CopyToRemote(remoteIP string, localPath string, remotePath string) error {
	sshCmd := exec.Command("scp", "-r", "-o", "StrictHostKeyChecking=no", localPath, fmt.Sprintf("root@%s:%s", remoteIP, remotePath))
	sshCmd.Stdout = os.Stdout
	sshCmd.Stderr = os.Stderr
	sshCmd.Stdin = os.Stdin

	err := sshCmd.Run()
	if err != nil {
		var exitError *exec.ExitError
		if !errors.As(err, &exitError) {
			return fmt.Errorf("unable to copy file to remote, ip address: %s, error: %w", remoteIP, err)
		}
	}

	return nil
}

// ExecuteSSHSession executes an SSH session to the given remote machine IP address.
func ExecuteSSHSession(remoteIP string) error {
	sshCmd := exec.Command("ssh", fmt.Sprintf("root@%s", remoteIP), "-o", "StrictHostKeyChecking=no")
	sshCmd.Stdout = os.Stdout
	sshCmd.Stderr = os.Stderr
	sshCmd.Stdin = os.Stdin

	err := sshCmd.Run()
	if err != nil {
		var exitError *exec.ExitError
		if !errors.As(err, &exitError) {
			return fmt.Errorf("unable start SSH session, ip address: %s, error: %w", remoteIP, err)
		}
	}

	return nil
}
