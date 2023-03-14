package ssh

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"go.uber.org/zap"
)

// ExecuteSSHSession executes an SSH session to the given VPS IP address.
func ExecuteSSHSession(logger *zap.SugaredLogger, vpsIPaddr string) error {
	sshCmd := exec.Command("ssh", fmt.Sprintf("root@%s", vpsIPaddr), "-o", "StrictHostKeyChecking=no")
	sshCmd.Stdout = os.Stdout
	sshCmd.Stderr = os.Stderr
	sshCmd.Stdin = os.Stdin

	err := sshCmd.Run()
	if err != nil {
		var exitError *exec.ExitError
		if !errors.As(err, &exitError) {
			logger.Errorw("Unable to start SSH session",
				"error", err,
				"VPS IP address", vpsIPaddr,
			)

			return err
		}
	}

	return nil
}

func CopyToRemote(logger *zap.SugaredLogger, vpsIPaddr string, localPath string, remotePath string) error {
	sshCmd := exec.Command("scp", "-r", "-o", "StrictHostKeyChecking=no", localPath, fmt.Sprintf("root@%s:%s", vpsIPaddr, remotePath))
	sshCmd.Stdout = os.Stdout
	sshCmd.Stderr = os.Stderr
	sshCmd.Stdin = os.Stdin

	err := sshCmd.Run()
	if err != nil {
		var exitError *exec.ExitError
		if !errors.As(err, &exitError) {
			logger.Errorw("Unable to copy file to remote",
				"error", err,
				"VPS IP address", vpsIPaddr,
			)

			return err
		}
	}

	return nil
}
