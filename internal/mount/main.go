package mount

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"strings"
)

const versionOutput string = "Amazon Simple Storage Service File System"

func NewMounter(mounter, mounterBinaryPath string) (Mounter, error) {
	switch mounter {
	case "s3fs":
		return S3fs{path: mounterBinaryPath}, nil
	default:
		return nil, fmt.Errorf("unknow mounter: %s", mounter)
	}
}

type Mounter interface {
	IsReady() (bool, error)
}

type S3fs struct {
	path string
}

func (s S3fs) IsReady() (bool, error) {
	cmd := exec.Command(s.path, "--version")
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return false, err
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return false, err
	}
	if err := cmd.Start(); err != nil {
		return false, err
	}
	stdout, err := ioutil.ReadAll(stdoutPipe)
	if err != nil {
		return false, err
	}
	stderr, err := ioutil.ReadAll(stderrPipe)
	if err != nil {
		return false, err
	}
	if err := cmd.Wait(); err != nil {
		// Check whether it is an error from running s3fs in which case return the error received via pipe
		if _, ok := err.(*exec.ExitError); ok {
			return false, fmt.Errorf("could not run %s --version: %v", s.path, stderr)
		}
	}

	// Check if the output is as expected
	if c := strings.Contains(string(stdout), versionOutput); !c {
		return false, fmt.Errorf("Unexpected %s --version output: %s", s.path, string(stdout))
	}
	return true, nil
}
