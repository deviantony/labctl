package ssh

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"go.uber.org/zap"
)

// CopyToRemote copies a file or directory to a remote flask
func CopyToRemote(logger *zap.SugaredLogger, flaskIP string, localPath string, remotePath string) error {
	sshCmd := exec.Command("scp", "-r", "-o", "StrictHostKeyChecking=no", localPath, fmt.Sprintf("root@%s:%s", flaskIP, remotePath))
	sshCmd.Stdout = os.Stdout
	sshCmd.Stderr = os.Stderr
	sshCmd.Stdin = os.Stdin

	err := sshCmd.Run()
	if err != nil {
		var exitError *exec.ExitError
		if !errors.As(err, &exitError) {
			logger.Errorw("Unable to copy file to remote",
				"error", err,
				"flask IP", flaskIP,
			)

			return err
		}
	}

	return nil
}

// ExecuteSSHSession executes an SSH session to the given flask IP address.
func ExecuteSSHSession(logger *zap.SugaredLogger, flaskIP string) error {
	sshCmd := exec.Command("ssh", fmt.Sprintf("root@%s", flaskIP), "-o", "StrictHostKeyChecking=no")
	sshCmd.Stdout = os.Stdout
	sshCmd.Stderr = os.Stderr
	sshCmd.Stdin = os.Stdin

	err := sshCmd.Run()
	if err != nil {
		var exitError *exec.ExitError
		if !errors.As(err, &exitError) {
			logger.Errorw("Unable to start SSH session",
				"error", err,
				"flask IP", flaskIP,
			)

			return err
		}
	}

	return nil
}
