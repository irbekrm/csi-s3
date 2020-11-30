package mount

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/google/fscrypt/filesystem"
)

const versionOutput string = "Amazon Simple Storage Service File System"

func NewMounter(mounter, mounterBinaryPath string) (Mounter, error) {
	switch mounter {
	case "s3fs":
		return s3fs{path: mounterBinaryPath}, nil
	default:
		return nil, fmt.Errorf("unknow mounter: %s", mounter)
	}
}

type Mounter interface {
	IsReady() (bool, error)
	Mount(string, string, string, string) error
	Unmount(string) (bool, error)
}

type s3fs struct {
	path string
}

// IsReady checks if s3fs binary is installed and valid
func (s s3fs) IsReady() (bool, error) {
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

func (s s3fs) Mount(mountPath, bucket, accessKey, secret string) error {
	// Storage provider is responsible for creating the directory
	err := os.Mkdir(mountPath, os.ModePerm)
	if os.IsExist(err) {
		// TODO: check if the right volume is mounted there
		return nil
	}
	if err != nil {
		return err
	}

	cmd := exec.Command(s.path, bucket, mountPath)
	// ensure the s3fs can read aws creds from env
	keyKV, secretKV := awsEnvVarsKV(accessKey, secret)
	cmd.Env = append(os.Environ(), keyKV, secretKV)
	return cmd.Run()
}

// UnMount idempotently unmounts filesystem mounted at targetPath and deletes the targetPath directory
func (s s3fs) Unmount(targetPath string) (bool, error) {
	found := false
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		// targetPath does not exist
		return found, nil
	}
	fss, err := filesystem.AllFilesystems()
	if err != nil {
		return found, fmt.Errorf("failed listing mounted filesystems: %v", err)
	}
	for _, fs := range fss {
		// found filesystem mounted at targetPath
		if fs.Path == targetPath {
			found = true
			err := syscall.Unmount(targetPath, 0)
			if err != nil {
				return found, fmt.Errorf("failed unmounting filesystem mounted at %s: %v", targetPath, err)
			}
		}
	}
	// delete targetPath directory
	if err := os.Remove(targetPath); err != nil {
		return found, fmt.Errorf("failed removing target path %s: %v", targetPath, err)
	}
	return found, nil
}

func awsEnvVarsKV(accessKey, secret string) (string, string) {
	return fmt.Sprintf("AWSACCESSKEYID=%s", accessKey), fmt.Sprintf("AWSSECRETACCESSKEY=%s", secret)
}
