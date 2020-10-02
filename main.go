package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"

	csi "github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/irbekrm/csi-s3/internal/mount"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

const (
	driverName = "s3.csi.irbe.dev"
	driverRepo = "https://github.com/irbekrm/csi-s3.git"
)

func main() {
	var (
		csiAddress        string
		driverVersion     string
		mounter           string
		mounterBinaryPath string
	)
	flag.StringVar(&csiAddress, "csi-address", "/csi/csi.sock", "Path of the UDS on which the gRPC server will serve Identity, Node, Controller services")
	flag.StringVar(&driverVersion, "v", "test", "driver release version")
	flag.StringVar(&mounter, mounter, "s3fs", "Mount backend. Currently only s3fs is supported")
	flag.StringVar(&mounterBinaryPath, "mounterBinaryPath", "/usr/local/bin/s3fs", "Path to the selected mount backend binary")
	flag.Parse()

	if err := os.RemoveAll(csiAddress); err != nil {
		fmt.Fprintf(os.Stderr, "could not remove socket %s: %v", csiAddress, err)
	}
	l, err := net.Listen("unix", csiAddress)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not listen on %s: %v", csiAddress, err)
		os.Exit(1)
	}
	defer l.Close()
	m, err := mount.NewMounter(mounter, mounterBinaryPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to set up mount backend: %v", err)
	}
	s := grpc.NewServer()
	i := identityServer{driverVersion: driverVersion, mounter: m}
	csi.RegisterIdentityServer(s, &i)
	reflection.Register(s)

	if err := s.Serve(l); err != nil {
		fmt.Fprintf(os.Stderr, "failed to run grpc server: %v", err)
		os.Exit(1)
	}
}

type identityServer struct {
	driverVersion string
	mounter       mount.Mounter
}

// GetPluginInfo returns information about this CSI plugin
func (s identityServer) GetPluginInfo(context.Context, *csi.GetPluginInfoRequest) (*csi.GetPluginInfoResponse, error) {
	m := map[string]string{"url": driverRepo}
	return &csi.GetPluginInfoResponse{
		Name:          driverName,
		VendorVersion: s.driverVersion,
		Manifest:      m,
	}, nil
}

func (s identityServer) GetPluginCapabilities(context.Context, *csi.GetPluginCapabilitiesRequest) (*csi.GetPluginCapabilitiesResponse, error) {
	return &csi.GetPluginCapabilitiesResponse{}, nil
}

func (s identityServer) Probe(context.Context, *csi.ProbeRequest) (*csi.ProbeResponse, error) {
	r, err := s.mounter.IsReady()
	if err != nil {
		err = status.Error(codes.FailedPrecondition, err.Error())
	}
	return &csi.ProbeResponse{Ready: &wrappers.BoolValue{Value: r}}, err
}
