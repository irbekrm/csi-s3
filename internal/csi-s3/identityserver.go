package csis3

import (
	"context"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/irbekrm/csi-s3/internal/mount"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	driverName = "s3.csi.irbe.dev"
	driverRepo = "https://github.com/irbekrm/csi-s3.git"
)

// NewIdentityServer returns a csi.IdentityServer implementation
func NewIdentityServer(driverVersion string, mounter mount.Mounter) csi.IdentityServer {
	return &identityServer{driverVersion, mounter}
}

type identityServer struct {
	driverVersion string
	mounter       mount.Mounter
}

// GetPluginInfo returns information about this CSI plugin
func (s *identityServer) GetPluginInfo(ctx context.Context, r *csi.GetPluginInfoRequest) (*csi.GetPluginInfoResponse, error) {
	m := map[string]string{"url": driverRepo}
	return &csi.GetPluginInfoResponse{
		Name:          driverName,
		VendorVersion: s.driverVersion,
		Manifest:      m,
	}, nil
}

// GetPluginCapabilities advertizes what non-default plugin capabilities this plugin has
func (s *identityServer) GetPluginCapabilities(ctx context.Context, r *csi.GetPluginCapabilitiesRequest) (*csi.GetPluginCapabilitiesResponse, error) {
	return &csi.GetPluginCapabilitiesResponse{}, nil
}

// Probe checks whether the plugin is functioning
func (s *identityServer) Probe(ctx context.Context, r *csi.ProbeRequest) (*csi.ProbeResponse, error) {
	ready, err := s.mounter.IsReady()
	if err != nil {
		err = status.Error(codes.FailedPrecondition, err.Error())
	}
	return &csi.ProbeResponse{Ready: &wrappers.BoolValue{Value: ready}}, err
}
