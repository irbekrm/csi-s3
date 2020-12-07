package main

import (
	"flag"
	"fmt"
	"net"
	"os"

	csi "github.com/container-storage-interface/spec/lib/go/csi"
	csis3 "github.com/irbekrm/csi-s3/internal/csi-s3"
	"github.com/irbekrm/csi-s3/internal/filesystem"
	"github.com/irbekrm/csi-s3/internal/mount"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
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

	m, err := mount.New(mounter, mounterBinaryPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to set up mount backend: %v", err)
	}
	fs := filesystem.New()

	s := grpc.NewServer()

	// register CSI Identity service
	i := csis3.NewIdentityServer(driverVersion, m)
	csi.RegisterIdentityServer(s, i)

	// register CSI Node service
	n := csis3.NewNodeServer(m, fs)
	csi.RegisterNodeServer(s, n)

	// For debugging purposes register reflection service
	reflection.Register(s)

	if err := s.Serve(l); err != nil {
		fmt.Fprintf(os.Stderr, "failed to run grpc server: %v", err)
		os.Exit(1)
	}
}
