package mount

//go:generate mockgen -source=main.go -destination=../../mocks/mock_mount.go -package=mocks
import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

const (
	versionOutput      string = "Amazon Simple Storage Service File System"
	fsType             string = "fuse.s3fs"
	envVarAwsAccessKey string = "AWSACCESSKEYID"
	envVarAwsSecretKey string = "AWSSECRETACCESSKEY"
)

func NewMounter(mounter, mounterBinaryPath string) (Mounter, error) {
	switch mounter {
	case "s3fs":
		return s3fs{mounterBinaryPath, run}, nil
	default:
		return nil, fmt.Errorf("unknow mounter: %s", mounter)
	}
}

type Mounter interface {
	IsReady() (bool, error)
	Mount(string, string, string, string, bool) error
	Type() string
}

type s3fs struct {
	path string
	run  func(cmd *exec.Cmd) (string, string, error)
}

// IsReady checks if s3fs binary is installed and valid
func (s s3fs) IsReady() (bool, error) {
	cmd := exec.Command(s.path, "--version")
	stdout, stderr, err := s.run(cmd)
	if err != nil {
		// Check whether it is an error from running s3fs in which case append stderr
		if _, ok := err.(*exec.ExitError); ok {
			return false, errors.Wrap(err, fmt.Sprintf("failed running %s: %s", s.path, stderr))
		}
		return false, errors.Wrap(err, "failed running command")
	}
	// Check if the output is as expected
	if c := strings.Contains(string(stdout), versionOutput); !c {
		return false, fmt.Errorf("Unexpected %s --version output: %s", s.path, string(stdout))
	}
	return true, nil
}

// Mount mounts bucket at the given path
// accessKey and secretKey are used to authenticate with AWS
// readonly determines if the mounted filesystem will be readonly
func (s s3fs) Mount(path, bucket, accessKey, secretKey string, readonly bool) error {
	cmd := exec.Command(s.path, bucket, path)
	// ensure the s3fs can read aws creds from env
	keyKV, secretKV := awsEnvVarsKV(accessKey, secretKey)
	cmd.Env = append(os.Environ(), keyKV, secretKV)
	_, stderr, err := s.run(cmd)
	if err != nil {
		// Check whether it is an error from running s3fs in which case append stderr
		if _, ok := err.(*exec.ExitError); ok {
			return errors.Wrap(err, fmt.Sprintf("failed running %s: %s", s.path, stderr))
		}
		return errors.Wrap(err, "failed running command")
	}
	return nil
}

// Type returns type name of filesystems s3fs creates
func (s3fs) Type() string {
	return fsType
}

func awsEnvVarsKV(accessKey, secret string) (string, string) {
	return fmt.Sprintf("%s=%s", envVarAwsAccessKey, accessKey), fmt.Sprintf("%s=%s", envVarAwsSecretKey, secret)
}

func run(cmd *exec.Cmd) (string, string, error) {
	var runErr error
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		runErr = errors.Wrap(err, "failed creating stdout pipe")
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		runErr = errors.Wrap(err, "failed creating stderr pipe")
	}
	if err != nil {
		return "", "", runErr
	}
	if err := cmd.Start(); err != nil {
		runErr = errors.Wrap(err, "failed starting the command")
	}
	stdout, err := ioutil.ReadAll(stdoutPipe)
	if err != nil {
		runErr = errors.Wrap(err, "failed reading from command stdout")
	}
	stderr, err := ioutil.ReadAll(stderrPipe)
	if err != nil {
		runErr = errors.Wrap(err, "failed reading from command stderr")
	}
	if err := cmd.Wait(); err != nil {
		runErr = errors.Wrap(err, "failed running the command")
	}
	return string(stdout), string(stderr), runErr
}

func commandWithEnv(cmd *exec.Cmd, kv map[string]string) *exec.Cmd {
	for k, v := range kv {
		cmd.Env = append(os.Environ(), fmt.Sprintf("%s=%s", k, v))
	}
	return cmd
}
