package util

import (
	"bytes"
	"io/ioutil"
	"os/exec"
	"strings"

	"github.com/develar/errors"
)

func IsWSL() bool {
	if GetCurrentOs() != LINUX {
		return false
	}

	release, err := getOsRelease()
	if err != nil {
		return false
	}

	if strings.Contains(strings.ToLower(release), "microsoft") {
		return true
	}

	version, err := getProcVersion()
	if err != nil {
		return false
	}

	if strings.Contains(strings.ToLower(version), "microsoft") {
		return true
	}

	return false
}

func getOsRelease() (string, error) {
	cmd := exec.Command("uname", "-r")

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", err
	}

	return out.String(), nil
}

func getProcVersion() (string, error) {
	content, err := ioutil.ReadFile("/proc/version")
	if err != nil {
		return "", errors.WithStack(err)
	}

	return string(content), nil
}
