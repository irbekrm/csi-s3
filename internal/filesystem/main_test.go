package filesystem_test

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	externalfs "github.com/google/fscrypt/filesystem"
	"github.com/irbekrm/csi-s3/internal/filesystem"
	"github.com/irbekrm/csi-s3/mocks"
)

func Test_fs_FindMount(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		want    filesystem.Matcher
		setup   func(ctrl *gomock.Controller, path string) filesystem.Sys
		wantErr bool
	}{
		{
			name: "target path does not exist",
			setup: func(ctrl *gomock.Controller, path string) filesystem.Sys {
				sys := mocks.NewMockSys(ctrl)
				sys.
					EXPECT().
					Stat(path).
					Return(nil, os.ErrNotExist)
				return sys
			},
		},
		{
			name: "fails retrieving fileinfo",
			setup: func(ctrl *gomock.Controller, path string) filesystem.Sys {
				sys := mocks.NewMockSys(ctrl)
				sys.
					EXPECT().
					Stat(path).
					Return(nil, errors.New("some error"))
				return sys
			},
			wantErr: true,
		},
		{
			name: "mountpoint not found",
			setup: func(ctrl *gomock.Controller, path string) filesystem.Sys {
				sys := mocks.NewMockSys(ctrl)
				sys.
					EXPECT().
					Stat(path).
					Return(nil, nil)
				sys.
					EXPECT().
					GetMount(path).
					Return(nil, fmt.Errorf("%s is not a mountpoint", path))
				return sys
			},
		},
		{
			name: "fails retrieving mount info",
			setup: func(ctrl *gomock.Controller, path string) filesystem.Sys {
				sys := mocks.NewMockSys(ctrl)
				sys.
					EXPECT().
					Stat(path).
					Return(nil, nil)
				sys.
					EXPECT().
					GetMount(path).
					Return(nil, errors.New("some error"))
				return sys
			},
			wantErr: true,
		},
		{
			name: "finds a mount",
			setup: func(ctrl *gomock.Controller, path string) filesystem.Sys {
				sys := mocks.NewMockSys(ctrl)
				sys.
					EXPECT().
					Stat(path).
					Return(nil, nil)
				sys.
					EXPECT().
					GetMount(path).
					Return(&externalfs.Mount{ReadOnly: true, FilesystemType: "some type"}, nil)
				return sys
			},
			want: filesystem.NewMatcher(true, "some type"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			sys := tt.setup(ctrl, tt.path)
			f := filesystem.New(filesystem.WithSys(sys))
			got, err := f.FindMount(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("fs.FindMount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("fs.FindMount() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_fs_EnsureMountRemoved(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		setup     func(ctrl *gomock.Controller, path string, err error) filesystem.Sys
		wantedErr error
		wantErr   bool
	}{
		{
			name: "target path does not exist",
			setup: func(ctrl *gomock.Controller, path string, err error) filesystem.Sys {
				sys := mocks.NewMockSys(ctrl)
				sys.
					EXPECT().
					Stat(path).
					Return(nil, os.ErrNotExist)
				return sys
			},
		},
		{
			name: "fails retrieving fileinfo",
			setup: func(ctrl *gomock.Controller, path string, err error) filesystem.Sys {
				sys := mocks.NewMockSys(ctrl)
				sys.
					EXPECT().
					Stat(path).
					Return(nil, errors.New("some error"))
				return sys
			},
			wantErr: true,
		},
		{
			name:      "mountpoint not found, fails to remove dir",
			wantedErr: errors.New("some error"),
			setup: func(ctrl *gomock.Controller, path string, err error) filesystem.Sys {
				sys := mocks.NewMockSys(ctrl)
				sys.
					EXPECT().
					Stat(path).
					Return(nil, nil)
				sys.
					EXPECT().
					GetMount(path).
					Return(nil, fmt.Errorf("%s is not a mountpoint", path))
				sys.
					EXPECT().
					Remove(path).
					Return(err)
				return sys
			},
			wantErr: true,
		},
		{
			name:      "mountpoint not found, success",
			wantedErr: errors.New("some error"),
			setup: func(ctrl *gomock.Controller, path string, err error) filesystem.Sys {
				sys := mocks.NewMockSys(ctrl)
				sys.
					EXPECT().
					Stat(path).
					Return(nil, nil)
				sys.
					EXPECT().
					GetMount(path).
					Return(nil, fmt.Errorf("%s is not a mountpoint", path))
				sys.
					EXPECT().
					Remove(path).
					Return(nil)
				return sys
			},
		},
		{
			name: "fails retrieving mount info",
			setup: func(ctrl *gomock.Controller, path string, err error) filesystem.Sys {
				sys := mocks.NewMockSys(ctrl)
				sys.
					EXPECT().
					Stat(path).
					Return(nil, nil)
				sys.
					EXPECT().
					GetMount(path).
					Return(nil, errors.New("some error"))
				return sys
			},
			wantErr: true,
		},
		{
			name:      "mount found, fails to unmount",
			wantedErr: errors.New("some error"),
			setup: func(ctrl *gomock.Controller, path string, err error) filesystem.Sys {
				sys := mocks.NewMockSys(ctrl)
				sys.
					EXPECT().
					Stat(path).
					Return(nil, nil)
				sys.
					EXPECT().
					GetMount(path).
					Return(nil, nil)
				sys.
					EXPECT().
					Unmount(path).
					Return(err)
				return sys
			},
			wantErr: true,
		},
		{
			name:      "mount found, successfully unmounted, fails to remove dir",
			wantedErr: errors.New("some error"),
			setup: func(ctrl *gomock.Controller, path string, err error) filesystem.Sys {
				sys := mocks.NewMockSys(ctrl)
				sys.
					EXPECT().
					Stat(path).
					Return(nil, nil)
				sys.
					EXPECT().
					GetMount(path).
					Return(nil, nil)
				sys.
					EXPECT().
					Unmount(path).
					Return(nil)
				sys.
					EXPECT().
					Remove(path).
					Return(err)
				return sys
			},
			wantErr: true,
		},
		{
			name: "mount found, successfully unmounted and removed dir",
			setup: func(ctrl *gomock.Controller, path string, err error) filesystem.Sys {
				sys := mocks.NewMockSys(ctrl)
				sys.
					EXPECT().
					Stat(path).
					Return(nil, nil)
				sys.
					EXPECT().
					GetMount(path).
					Return(nil, nil)
				sys.
					EXPECT().
					Unmount(path).
					Return(nil)
				sys.
					EXPECT().
					Remove(path).
					Return(err)
				return sys
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			sys := tt.setup(ctrl, tt.path, tt.wantedErr)
			f := filesystem.New(filesystem.WithSys(sys))
			if err := f.EnsureMountRemoved(tt.path); (err != nil) != tt.wantErr {
				t.Errorf("fs.EnsureMountRemoved() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_fs_EnsureDirExists(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		finfo     os.FileInfo
		setup     func(ctrl *gomock.Controller, path string, err error, finfo os.FileInfo) filesystem.Sys
		wantedErr error
		wantErr   bool
	}{
		{
			name:      "dir does not exist, fails making it",
			wantedErr: errors.New("some error"),
			setup: func(ctrl *gomock.Controller, path string, err error, finfo os.FileInfo) filesystem.Sys {
				sys := mocks.NewMockSys(ctrl)
				sys.
					EXPECT().
					Stat(path).
					Return(finfo, os.ErrNotExist)
				sys.
					EXPECT().
					Mkdir(path, os.ModePerm).
					Return(err)
				return sys
			},
			wantErr: true,
		},
		{
			name: "dir does not exist, successfully creates it",
			setup: func(ctrl *gomock.Controller, path string, err error, finfo os.FileInfo) filesystem.Sys {
				sys := mocks.NewMockSys(ctrl)
				sys.
					EXPECT().
					Stat(path).
					Return(finfo, os.ErrNotExist)
				sys.
					EXPECT().
					Mkdir(path, os.ModePerm).
					Return(err)
				return sys
			},
		},
		{
			name:      "fails retrieving directory info",
			wantedErr: errors.New("some error"),
			setup: func(ctrl *gomock.Controller, path string, err error, finfo os.FileInfo) filesystem.Sys {
				sys := mocks.NewMockSys(ctrl)
				sys.
					EXPECT().
					Stat(path).
					Return(finfo, err)
				return sys
			},
			wantErr: true,
		},
		{
			name: "directory already exists at the given path",
			setup: func(ctrl *gomock.Controller, path string, err error, finfo os.FileInfo) filesystem.Sys {
				sys := mocks.NewMockSys(ctrl)
				sys.
					EXPECT().
					Stat(path).
					Return(finfo, err)
				sys.
					EXPECT().
					IsDir(finfo).
					Return(true)
				return sys
			},
		},
		{
			name: "some file already exists at the given path, but is not a directory",
			setup: func(ctrl *gomock.Controller, path string, err error, finfo os.FileInfo) filesystem.Sys {
				sys := mocks.NewMockSys(ctrl)
				sys.
					EXPECT().
					Stat(path).
					Return(finfo, err)
				sys.
					EXPECT().
					IsDir(finfo).
					Return(false)
				return sys
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			sys := tt.setup(ctrl, tt.path, tt.wantedErr, tt.finfo)
			f := filesystem.New(filesystem.WithSys(sys))
			if err := f.EnsureDirExists(tt.path); (err != nil) != tt.wantErr {
				t.Errorf("fs.EnsureDirExists() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
