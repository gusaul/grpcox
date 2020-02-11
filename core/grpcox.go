package core

import (
	"context"
	"os"
	"reflect"
	"strconv"
	"time"

	"github.com/fullstorydev/grpcurl"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

// GrpCox - main object
type GrpCox struct {
	KeepAlive float64

	activeConn  *ConnStore
	maxLifeConn time.Duration

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

// Proto define protofile uploaded from client
// will be used to be persisted to disk and indicator
// whether connections should reflect from server or local proto
type Proto struct {
	Name    string
	Content []byte
}

// InitGrpCox constructor
func InitGrpCox() *GrpCox {
	maxLife, tick := 10, 3

	if val, err := strconv.Atoi(os.Getenv("MAX_LIFE_CONN")); err == nil {
		maxLife = val
	}

	if val, err := strconv.Atoi(os.Getenv("TICK_CLOSE_CONN")); err == nil {
		tick = val
	}

	c := NewConnectionStore()
	g := &GrpCox{
		activeConn: c,
	}

	if maxLife > 0 && tick > 0 {
		g.maxLifeConn = time.Duration(maxLife) * time.Minute
		c.StartGC(time.Duration(tick) * time.Second)
	}

	return g
}

// GetResource - open resource to targeted grpc server
func (g *GrpCox) GetResource(ctx context.Context, target string, plainText, isRestartConn bool) (*Resource, error) {
	if conn, ok := g.activeConn.getConnection(target); ok {
		if !isRestartConn && conn.isValid() {
			return conn, nil
		}
		g.CloseActiveConns(target)
	}

	var err error
	r := new(Resource)
	h := append(g.headers, g.reflectHeaders...)
	r.md = grpcurl.MetadataFromHeaders(h)
	r.clientConn, err = g.dial(ctx, target, plainText)
	if err != nil {
		return nil, err
	}

	r.headers = h

	g.activeConn.addConnection(target, r, g.maxLifeConn)
	return r, nil
}

// GetResourceWithProto - open resource to targeted grpc server using given protofile
func (g *GrpCox) GetResourceWithProto(ctx context.Context, target string, plainText, isRestartConn bool, protos []Proto) (*Resource, error) {
	r, err := g.GetResource(ctx, target, plainText, isRestartConn)
	if err != nil {
		return nil, err
	}

	// if given protofile is equal to current, skip adding protos as it's already
	// persisted in the harddisk anyway
	if reflect.DeepEqual(r.protos, protos) {
		return r, nil
	}

	// add protos property to resource and persist it to harddisk
	err = r.AddProtos(protos)
	return r, err
}

// GetActiveConns - get all saved active connection
func (g *GrpCox) GetActiveConns(ctx context.Context) []string {
	active := g.activeConn.getAllConn()
	result := make([]string, len(active))
	i := 0
	for k := range active {
		result[i] = k
		i++
	}
	return result
}

// CloseActiveConns - close conn by host or all
func (g *GrpCox) CloseActiveConns(host string) error {
	if host == "all" {
		for k := range g.activeConn.getAllConn() {
			g.activeConn.delete(k)
		}
		return nil
	}

	g.activeConn.delete(host)
	return nil
}

// Extend extend connection based on setting max life
func (g *GrpCox) Extend(host string) {
	g.activeConn.extend(host, g.maxLifeConn)
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
