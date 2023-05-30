package prompt

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// AskFor2FACode asks the user for a 2FA code.
func AskFor2FACode() (string, error) {
	fmt.Print("Enter 2FA code: ")
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return "", err
	}

	code := ""
	for i := 0; i < 6; i++ {
		b := make([]byte, 1)
		_, err = os.Stdin.Read(b)
		if err != nil {
			return "", err
		}

		codeChar := string(b)
		fmt.Print(codeChar)
		code += codeChar
	}

	err = term.Restore(int(os.Stdin.Fd()), oldState)
	if err != nil {
		return "", err
	}

	fmt.Println()
	return code, nil
}

// AskForConfirmation asks the user for a confirmation.
func AskForConfirmation() (bool, error) {
	r := bufio.NewReader(os.Stdin)
	line, _, err := r.ReadLine()
	if err != nil {
		return false, err
	}

	response := string(line)

	switch strings.ToLower(response) {
	case "y", "yes":
		return true, nil
	case "n", "no":
		return false, nil
	case "":
		return false, nil
	default:
		fmt.Println("Please type [y]es or [n]o and then press enter:")
		return AskForConfirmation()
	}
}
