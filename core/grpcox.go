package core

import (
	"context"
	"time"

	"github.com/fullstorydev/grpcurl"
	"github.com/jhump/protoreflect/grpcreflect"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
	reflectpb "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
)

// GrpCox - main object
type GrpCox struct {
	KeepAlive float64
	PlainText bool

	// TODO : utilize below args
	headers        []string
	reflectHeaders []string
	authority      string
	insecure       bool
	cacert         string
	cert           string
	key            string
	serverName     string
	isUnixSocket   func() bool
}

// GetResource - open resource to targeted grpc server
func (g *GrpCox) GetResource(ctx context.Context, target string) (*Resource, error) {
	var err error
	r := new(Resource)
	h := append(g.headers, g.reflectHeaders...)
	md := grpcurl.MetadataFromHeaders(h)
	refCtx := metadata.NewOutgoingContext(ctx, md)
	r.clientConn, err = g.dial(ctx, target)
	if err != nil {
		return nil, err
	}

	r.refClient = grpcreflect.NewClient(refCtx, reflectpb.NewServerReflectionClient(r.clientConn))
	r.descSource = grpcurl.DescriptorSourceFromServer(ctx, r.refClient)
	r.headers = h

	return r, nil
}

func (g *GrpCox) dial(ctx context.Context, target string) (*grpc.ClientConn, error) {
	dialTime := 10 * time.Second
	ctx, cancel := context.WithTimeout(ctx, dialTime)
	defer cancel()
	var opts []grpc.DialOption

	// keep alive
	if g.KeepAlive > 0 {
		timeout := time.Duration(g.KeepAlive * float64(time.Second))
		opts = append(opts, grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:    timeout,
			Timeout: timeout,
		}))
	}

	if g.authority != "" {
		opts = append(opts, grpc.WithAuthority(g.authority))
	}

	var creds credentials.TransportCredentials
	if !g.PlainText {
		var err error
		creds, err = grpcurl.ClientTransportCredentials(g.insecure, g.cacert, g.cert, g.key)
		if err != nil {
			return nil, err
		}
		if g.serverName != "" {
			if err := creds.OverrideServerName(g.serverName); err != nil {
				return nil, err
			}
		}
	}
	network := "tcp"
	if g.isUnixSocket != nil && g.isUnixSocket() {
		network = "unix"
	}
	cc, err := grpcurl.BlockingDial(ctx, network, target, creds, opts...)
	if err != nil {
		return nil, err
	}
	return cc, nil
}
