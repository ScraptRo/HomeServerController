package Internal_Process_Handler

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
)

type command_out struct {
	Output string `json:"out"`
	Err    string `json:"error"`
}
func RunScript(args []string) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("no script provided")
	}

	script := args[0]
	scriptArgs := args[1:]
	ext := filepath.Ext(script)

	var cmd *exec.Cmd

	switch ext {
	case ".py":
		if runtime.GOOS == "windows" {
			cmd = exec.Command("python", append([]string{script}, scriptArgs...)...)
		} else {
			cmd = exec.Command("python3", append([]string{script}, scriptArgs...)...)
		}
	case ".sh":
		cmd = exec.Command("sh", append([]string{script}, scriptArgs...)...)
	case ".bat":
		if runtime.GOOS != "windows" {
			return "", fmt.Errorf(".bat scripts only supported on Windows")
		}
		cmd = exec.Command("cmd", append([]string{"/C", script}, scriptArgs...)...)
	case ".ps1":
		if runtime.GOOS != "windows" {
			return "", fmt.Errorf(".ps1 scripts only supported on Windows")
		}
		cmd = exec.Command("powershell", append([]string{"-File", script}, scriptArgs...)...)
	default:
		// try to run as a normal executable
		cmd = exec.Command(script, scriptArgs...)
	}

	out, err := cmd.CombinedOutput()
	return string(out), err
}
func RunCommand(args []string) string {
	var cmdOut command_out
	if len(args) == 0 {
		cmdOut.Err = "no command provided"
		b, _ := json.Marshal(cmdOut)
		return string(b)
	}

	// Build command string
	cmdStr := ""
	for _, a := range args {
		if cmdStr == "" {
			cmdStr = a
		} else {
			cmdStr += " " + a
		}
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", cmdStr)
	} else {
		cmd = exec.Command("sh", "-c", cmdStr)
	}

	out, err := cmd.CombinedOutput()
	cmdOut.Output = string(out)
	if err != nil {
		cmdOut.Err = err.Error()
	}

	b, _ := json.Marshal(cmdOut)
	return string(b)
}
