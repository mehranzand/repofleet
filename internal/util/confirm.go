package util

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func Confirm(msg string) bool {
	fmt.Printf("  %s [y/N] ", msg)
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
		return answer == "y" || answer == "yes"
	}
	return false
}
