package csis3

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/mock/gomock"
	"github.com/irbekrm/csi-s3/internal/filesystem"
	"github.com/irbekrm/csi-s3/internal/mount"
	"github.com/irbekrm/csi-s3/mocks"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Test_nodeServer_NodePublishVolume(t *testing.T) {
	tests := []struct {
		name        string
		mounterType string
		readonly    bool
		in          *csi.NodePublishVolumeRequest
		setup       func(*gomock.Controller, string, bool) (mount.Mounter, filesystem.FS)
		want        *csi.NodePublishVolumeResponse
		RPCCode     codes.Code
		wantErr     bool
	}{
		{
			name: "fails looking for mount at targetpath",
			in:   &csi.NodePublishVolumeRequest{TargetPath: "some path"},
			setup: func(ctrl *gomock.Controller, mounterType string, readonly bool) (mount.Mounter, filesystem.FS) {
				fs := mocks.NewMockFS(ctrl)
				fs.
					EXPECT().
					FindMount("some path").
					Return(nil, errors.New("some error"))
				return nil, fs
			},
			want:    &csi.NodePublishVolumeResponse{},
			RPCCode: codes.Internal,
			wantErr: true,
		},
		{
			name:        "finds a non-matching mount at target path",
			in:          &csi.NodePublishVolumeRequest{TargetPath: "some path"},
			mounterType: "some type",
			setup: func(ctrl *gomock.Controller, mounterType string, readonly bool) (mount.Mounter, filesystem.FS) {
				matcher := mocks.NewMockMatcher(ctrl)
				matcher.
					EXPECT().
					Match(mounterType, readonly).
					Return(false)
				mounter := mocks.NewMockMounter(ctrl)
				mounter.
					EXPECT().
					Type().
					Return(mounterType)
				fs := mocks.NewMockFS(ctrl)
				fs.
					EXPECT().
					FindMount("some path").
					Return(matcher, nil)
				return mounter, fs
			},
			want:    &csi.NodePublishVolumeResponse{},
			RPCCode: codes.AlreadyExists,
			wantErr: true,
		},
		{
			name:        "nothing to do, finds a matching mount at target path",
			in:          &csi.NodePublishVolumeRequest{TargetPath: "some path"},
			mounterType: "some type",
			setup: func(ctrl *gomock.Controller, mounterType string, readonly bool) (mount.Mounter, filesystem.FS) {
				matcher := mocks.NewMockMatcher(ctrl)
				matcher.
					EXPECT().
					Match(mounterType, readonly).
					Return(true)
				mounter := mocks.NewMockMounter(ctrl)
				mounter.
					EXPECT().
					Type().
					Return(mounterType)
				fs := mocks.NewMockFS(ctrl)
				fs.
					EXPECT().
					FindMount("some path").
					Return(matcher, nil)
				return mounter, fs
			},
			want:    &csi.NodePublishVolumeResponse{},
			RPCCode: codes.OK,
		},
		{
			name:        "fails to create directory at target path",
			in:          &csi.NodePublishVolumeRequest{TargetPath: "some path"},
			mounterType: "some type",
			setup: func(ctrl *gomock.Controller, mounterType string, readonly bool) (mount.Mounter, filesystem.FS) {
				fs := mocks.NewMockFS(ctrl)
				fs.
					EXPECT().
					FindMount("some path").
					Return(nil, nil)
				fs.
					EXPECT().
					EnsureDirExists("some path").
					Return(errors.New("some error"))
				return nil, fs
			},
			want:    &csi.NodePublishVolumeResponse{},
			RPCCode: codes.Internal,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// set up mocks
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mnt, fs := tt.setup(ctrl, tt.mounterType, tt.readonly)

			n := &nodeServer{
				mounter: mnt,
				fs:      fs,
			}
			ctx := context.TODO()

			got, err := n.NodePublishVolume(ctx, tt.in)

			// Check the returned RPCCode
			gRPCStatus, ok := status.FromError(err)
			if !ok {
				t.Fatalf("expected RPC status code: %v got non-RPC error: %v", tt.RPCCode, err)
			}
			if tt.RPCCode != gRPCStatus.Code() {
				t.Fatalf("expected RPC status code: %v, got: %v\n", tt.RPCCode, gRPCStatus.Code())
			}

			if (err != nil) != tt.wantErr {
				t.Fatalf("nodeServer.NodePublishVolume() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("nodeServer.NodePublishVolume() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_nodeServer_NodeGetInfo(t *testing.T) {
	tests := []struct {
		name    string
		nodeId  string
		in      *csi.NodeGetInfoRequest
		want    *csi.NodeGetInfoResponse
		RPCCode codes.Code
		wantErr bool
	}{
		{
			name:    "success",
			nodeId:  "some id",
			want:    &csi.NodeGetInfoResponse{NodeId: "some id"},
			RPCCode: codes.OK,
		},
		{
			name:    "failure, node id not found",
			want:    &csi.NodeGetInfoResponse{},
			RPCCode: codes.Internal,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.TODO()
			n := &nodeServer{
				nodeId: tt.nodeId,
			}
			got, err := n.NodeGetInfo(ctx, tt.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("nodeServer.NodeGetInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("nodeServer.NodeGetInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}
