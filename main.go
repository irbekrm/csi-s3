package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
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
	c := controllerServer{}
	csi.RegisterControllerServer(s, c)
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
func (s identityServer) GetPluginInfo(ctx context.Context, r *csi.GetPluginInfoRequest) (*csi.GetPluginInfoResponse, error) {
	m := map[string]string{"url": driverRepo}
	return &csi.GetPluginInfoResponse{
		Name:          driverName,
		VendorVersion: s.driverVersion,
		Manifest:      m,
	}, nil
}

func (s identityServer) GetPluginCapabilities(ctx context.Context, r *csi.GetPluginCapabilitiesRequest) (*csi.GetPluginCapabilitiesResponse, error) {
	// CONTROLLER_SERVICE plugin capability
	cSvc := csi.PluginCapability{Type: &csi.PluginCapability_Service_{Service: &csi.PluginCapability_Service{Type: csi.PluginCapability_Service_CONTROLLER_SERVICE}}}
	caps := []*csi.PluginCapability{&cSvc}
	return &csi.GetPluginCapabilitiesResponse{Capabilities: caps}, nil
}

func (s identityServer) Probe(ctx context.Context, r *csi.ProbeRequest) (*csi.ProbeResponse, error) {
	ready, err := s.mounter.IsReady()
	if err != nil {
		err = status.Error(codes.FailedPrecondition, err.Error())
	}
	return &csi.ProbeResponse{Ready: &wrappers.BoolValue{Value: ready}}, err
}

type controllerServer struct {
	*csi.UnimplementedControllerServer
}

// ControllerGetCapabilities advertizes which capabilities this controller supports
func (s controllerServer) ControllerGetCapabilities(ctx context.Context, r *csi.ControllerGetCapabilitiesRequest) (*csi.ControllerGetCapabilitiesResponse, error) {
	cd := csi.ControllerServiceCapability{Type: &csi.ControllerServiceCapability_Rpc{Rpc: &csi.ControllerServiceCapability_RPC{Type: csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME}}}
	caps := []*csi.ControllerServiceCapability{&cd}
	return &csi.ControllerGetCapabilitiesResponse{Capabilities: caps}, nil
}

func (s controllerServer) CreateVolume(ctx context.Context, r *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {
	// retrieve AWS creds from csi.CreateVolumeRequest.Secrets
	key, secret, ok := awsCreds(r.Secrets)
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "iaas creds not provided")
	}
	bucket, ok := r.Parameters["bucket"]
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "bucket name not provided")
	}
	// TODO: do I need to validate these credentials?
	creds := credentials.NewStaticCredentials(key, secret, "")
	cfg := aws.NewConfig()
	cfg = cfg.WithCredentials(creds)
	// TODO: do I need to require for the region to be passed in params?
	cfg = cfg.WithRegion("eu-west-2")
	sess, err := session.NewSession(cfg)
	if err != nil {
		return nil, status.Error(codes.Internal, "could not create an aws session")
	}
	// TODO: Call list tags on the bucket to determine if it is there
	// TODO: Check the tags to see if it was created by this plugin? (Maybe no need to implement now, could just return in volume_context that it was not)
	svc := s3.New(sess)
	input := &s3.GetBucketTaggingInput{
		Bucket: aws.String(bucket),
	}
	_, err = svc.GetBucketTagging(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				fmt.Println("Error:", aerr.Error())
				return nil, status.Error(codes.InvalidArgument, "static bucket mounting requested, but bucket not found")
			}
		} else {
			fmt.Println("Non-aws error:", err)
			return nil, status.Error(codes.Internal, "error retrieving bucket tags: %v")
		}
	}
	return nil, nil
}

// TODO: move this whole thing to iaas (?) package and see if creds can be put into a struct or something
func awsCreds(s map[string]string) (string, string, bool) {
	key, keyOk := s["AWS_ACCESS_KEY_ID"]
	secret, secretOk := s["AWS_SECRET_ACCESS_KEY"]
	return key, secret, keyOk && secretOk
}
