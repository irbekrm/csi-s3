package csis3

import (
	"context"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/irbekrm/csi-s3/internal/filesystem"
	"github.com/irbekrm/csi-s3/internal/mount"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// NewNodeServer returns a csi.NodeServer implementation
func NewNodeServer(mounter mount.Mounter, fs filesystem.FS) csi.NodeServer {
	return &nodeServer{mounter: mounter, fs: fs}
}

type nodeServer struct {
	*csi.UnimplementedNodeServer
	mounter mount.Mounter
	fs      filesystem.FS
}

// NodePublishVolume mounts the volume at the specified path (in the container). Safe to be called multiple times
func (n *nodeServer) NodePublishVolume(ctx context.Context, in *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	// TODO: first verify that the bucket (volume_id) exists
	// check if a mount already exists at the targetPath
	targetPath := in.TargetPath
	m, err := n.fs.FindMount(targetPath)
	if err != nil {
		return &csi.NodePublishVolumeResponse{}, status.Error(codes.Internal, err.Error())
	}

	// if a mount already exists at targetPath, check that it's the right one
	//TODO: match volume_id
	readonly := in.Readonly
	if m != nil {
		ok := m.Match(n.mounter.Type(), readonly)
		if !ok {
			return &csi.NodePublishVolumeResponse{}, status.Error(codes.AlreadyExists, "")
		} else {
			return &csi.NodePublishVolumeResponse{}, status.Error(codes.OK, "")
		}
	}

	// mount does not yet exist, proceed
	if err := n.fs.EnsureDirExists(targetPath); err != nil {
		return &csi.NodePublishVolumeResponse{}, status.Error(codes.Internal, err.Error())
	}
	bucket := in.VolumeId
	// retrieve AWS creds from csi.NodePublishVolumeRequest.Secrets
	key, secret, ok := awsCreds(in.Secrets)
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "iaas creds not provided")
	}
	if err := n.mounter.Mount(targetPath, bucket, key, secret, false); err != nil {
		return &csi.NodePublishVolumeResponse{}, status.Error(codes.Internal, err.Error())
	}
	return &csi.NodePublishVolumeResponse{}, status.Error(codes.OK, "")
}

// NodeUnpublishVolume idempotently unmounts the volume from the given target path
func (n *nodeServer) NodeUnpublishVolume(ctx context.Context, in *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	// TODO: first verify that the bucket (volume_id) exists
	targetPath := in.TargetPath
	resp := &csi.NodeUnpublishVolumeResponse{}
	if err := n.fs.EnsureMountRemoved(targetPath); err != nil {
		return resp, status.Error(codes.Internal, err.Error())
	}
	return resp, status.Error(codes.OK, "")
}
