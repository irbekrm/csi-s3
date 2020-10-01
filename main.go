package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"

	csi "github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc"
)

func main() {
	var csiAddress string
	flag.StringVar(&csiAddress, "csi-address", "/csi/csi.sock", "Path of the UDS on which the gRPC server will serve Identity, Node, Controller services")
	flag.Parse()

	s := grpc.NewServer()
	i := identityServer{}
	csi.RegisterIdentityServer(s, &i)

	if err := os.RemoveAll(csiAddress); err != nil {
		fmt.Fprintf(os.Stderr, "could not remove socket %s: %v", csiAddress, err)
	}
	l, err := net.Listen("unix", csiAddress)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not listen on %s: %v", csiAddress, err)
		os.Exit(1)
	}
	defer l.Close()

	if err := s.Serve(l); err != nil {
		fmt.Fprintf(os.Stderr, "failed to run grpc server: %v", err)
		os.Exit(1)
	}
}

type identityServer struct{}

func (s identityServer) GetPluginInfo(context.Context, *csi.GetPluginInfoRequest) (*csi.GetPluginInfoResponse, error) {
	return &csi.GetPluginInfoResponse{}, nil
}

func (s identityServer) GetPluginCapabilities(context.Context, *csi.GetPluginCapabilitiesRequest) (*csi.GetPluginCapabilitiesResponse, error) {
	return &csi.GetPluginCapabilitiesResponse{}, nil
}

func (s identityServer) Probe(context.Context, *csi.ProbeRequest) (*csi.ProbeResponse, error) {
	return &csi.ProbeResponse{}, nil
}
