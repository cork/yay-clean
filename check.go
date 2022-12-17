package main

import (
	"os/exec"
)

func CheckInstalled(name string) bool {
	cmd := exec.Command("yay", "-Qi", name)

	cmd.Run()

	return cmd.ProcessState.ExitCode() == 0
}
