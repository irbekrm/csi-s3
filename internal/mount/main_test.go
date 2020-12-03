package mount

import (
	"errors"
	"os/exec"
	"testing"
)

func Test_s3fs_IsReady(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		run     func(*exec.Cmd) (string, string, error)
		want    bool
		wantErr bool
	}{
		{
			name: "failed executing command",
			run: func(cmd *exec.Cmd) (string, string, error) {
				return "", "", errors.New("some error")
			},
			wantErr: true,
		},
		{
			name: "did not find expected output",
			run: func(cmd *exec.Cmd) (string, string, error) {
				return "", "", nil
			},
			wantErr: true,
		},
		{
			name: "success",
			run: func(cmd *exec.Cmd) (string, string, error) {
				return versionOutput, "", nil
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := s3fs{
				path: tt.path,
				run:  tt.run,
			}
			got, err := s.IsReady()
			if (err != nil) != tt.wantErr {
				t.Errorf("s3fs.IsReady() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("s3fs.IsReady() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_s3fs_Mount(t *testing.T) {
	tests := []struct {
		name      string
		mountPath string
		bucket    string
		accessKey string
		secretKey string
		readonly  bool
		run       func(cmd *exec.Cmd) (string, string, error)
		wantErr   bool
	}{
		{
			name: "failed executing command",
			run: func(cmd *exec.Cmd) (string, string, error) {
				return "", "", errors.New("some error")
			},
			wantErr: true,
		},
		{
			name: "success",
			run: func(cmd *exec.Cmd) (string, string, error) {
				return "", "", nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := s3fs{
				run: tt.run,
			}
			if err := s.Mount(tt.mountPath, tt.bucket, tt.accessKey, tt.secretKey, tt.readonly); (err != nil) != tt.wantErr {
				t.Errorf("s3fs.Mount() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
