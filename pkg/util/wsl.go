package util

import (
	"bytes"
	"io/ioutil"
	"os/exec"
	"strings"
)

func IsWSL() bool {
	if GetCurrentOs() != LINUX {
		return false
	}

	err, release := getOSRelease()
	if err != nil {
		return false
	}

	if strings.Contains(strings.ToLower(release), "microsoft") {
		return true
	}

	err, version := getProcVersion()
	if err != nil {
		return false
	}

	if strings.Contains(strings.ToLower(version), "microsoft") {
		return true
	}

	return false
}

func getOSRelease() (error, string) {
	cmd := exec.Command("uname","-r")

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	
	err := cmd.Run()
	if err != nil {
		return err, ""
	}

	return nil, out.String()
}

func getProcVersion() (error, string) {
	content, err := ioutil.ReadFile("/proc/version")
	if err != nil {
		return err, ""
	}

	return nil, string(content)
}
