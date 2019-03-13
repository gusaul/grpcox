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

	activeConn map[string]*Resource

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

// InitGrpCox constructor
func InitGrpCox() *GrpCox {
	return &GrpCox{
		activeConn: make(map[string]*Resource),
	}
}

// GetResource - open resource to targeted grpc server
func (g *GrpCox) GetResource(ctx context.Context, target string, plainText, isRestartConn bool) (*Resource, error) {
	if conn, ok := g.activeConn[target]; ok {
		if !isRestartConn && conn.refClient != nil && conn.clientConn != nil {
			return conn, nil
		}
		g.CloseActiveConns(target)
	}

	var err error
	r := new(Resource)
	h := append(g.headers, g.reflectHeaders...)
	md := grpcurl.MetadataFromHeaders(h)
	refCtx := metadata.NewOutgoingContext(ctx, md)
	r.clientConn, err = g.dial(ctx, target, plainText)
	if err != nil {
		return nil, err
	}

	r.refClient = grpcreflect.NewClient(refCtx, reflectpb.NewServerReflectionClient(r.clientConn))
	r.descSource = grpcurl.DescriptorSourceFromServer(ctx, r.refClient)
	r.headers = h

	g.activeConn[target] = r
	return r, nil
}

// GetActiveConns - get all saved active connection
func (g *GrpCox) GetActiveConns(ctx context.Context) []string {
	result := make([]string, len(g.activeConn))
	i := 0
	for k := range g.activeConn {
		result[i] = k
		i++
	}
	return result
}

// CloseActiveConns - close conn by host or all
func (g *GrpCox) CloseActiveConns(host string) error {
	if host == "all" {
		for k, v := range g.activeConn {
			v.Close()
			delete(g.activeConn, k)
		}
		return nil
	}

	if v, ok := g.activeConn[host]; ok {
		v.Close()
		delete(g.activeConn, host)
	}

	return nil
}

func (g *GrpCox) dial(ctx context.Context, target string, plainText bool) (*grpc.ClientConn, error) {
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
	if !plainText {
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
