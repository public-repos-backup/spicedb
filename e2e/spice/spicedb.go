package spice

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/authzed/grpcutil"
	"github.com/authzed/spicedb/e2e"
	"google.golang.org/grpc"

	"github.com/authzed/spicedb/e2e/cockroach"
)

//go:generate go run github.com/ecordell/optgen -output spicedb_options.go . SpiceDb

// Node represents a single instance of spicedb started via exec
type Node struct {
	ID            string
	PresharedKey  string
	Datastore     string
	DbName        string
	URI           string
	GrpcPort      int
	InternalPort  int
	MetricsPort   int
	DashboardPort int
	Pid           int
	Cancel        context.CancelFunc
	client        e2e.Client
}

// WithTestDefaults sets the default values for Node
func WithTestDefaults(opts ...SpiceDbOption) SpiceDbOption {
	return func(s *Node) {
		for _, o := range opts {
			o(s)
		}
		if s.GrpcPort == 0 {
			s.GrpcPort = 50051
		}
		if s.InternalPort == 0 {
			s.InternalPort = 50052
		}
		if s.MetricsPort == 0 {
			s.MetricsPort = 9090
		}
		if s.DashboardPort == 0 {
			// this would typically be 8080, but conflicts with Node's dash
			s.DashboardPort = 8090
		}
		if len(s.DbName) == 0 {
			s.DbName = "spicedb"
		}
		if len(s.PresharedKey) == 0 {
			s.PresharedKey = "testtesttesttest"
		}
	}
}

// Start starts an instance of spicedb using the configuration
func (s *Node) Start(ctx context.Context, logprefix string, args ...string) error {
	logfile, err := e2e.File(ctx, fmt.Sprintf("spicedb-%s-%s.log", logprefix, s.ID))
	if err != nil {
		return err
	}
	cmd := []string{
		"./spicedb",
		"serve",
		"--log-level=debug",
		"--grpc-preshared-key=" + s.PresharedKey,
		"--grpc-no-tls",
		"--datastore-engine=" + s.Datastore,
		"--datastore-conn-uri=" + s.URI,
		fmt.Sprintf("--grpc-addr=:%d", s.GrpcPort),
		fmt.Sprintf("--internal-grpc-addr=:%d", s.InternalPort),
		fmt.Sprintf("--metrics-addr=:%d", s.MetricsPort),
		fmt.Sprintf("--dashboard-addr=:%d", s.DashboardPort),
	}

	ctx, cancel := context.WithCancel(ctx)
	s.Cancel = cancel
	s.Pid, err = e2e.GoRun(ctx, logfile, logfile, append(cmd, args...)...)
	return err
}

// Stop will Cancel a running spicedb process
func (s *Node) Stop() error {
	if s.Pid < 1 {
		return fmt.Errorf("can't stop an unstarted spicedb")
	}
	s.Cancel()
	return nil
}

// Connect blocks until a connection to the spicedb instance can be established.
// Once connected, the client is avaialable via Client()
func (s *Node) Connect(ctx context.Context, out io.Writer) error {
	if s.Pid < 1 {
		return fmt.Errorf("can't create client for unstarted spicedb")
	}

	addr := net.JoinHostPort("localhost", strconv.Itoa(s.GrpcPort))
	e2e.WaitForServerReady(addr, out)

	conn, err := grpc.DialContext(ctx, addr,
		grpc.WithBlock(), grpc.WithInsecure(),
		grpcutil.WithInsecureBearerToken(s.PresharedKey))
	if err != nil {
		return err
	}
	s.client = e2e.NewClient(conn)
	return nil
}

// Client returns a client that can talk to a started spicedb instance
func (s *Node) Client() e2e.Client {
	return s.client
}

// Cluster is a set of spicedb nodes
type Cluster []*Node

// NewClusterFromCockroachCluster creates a spicedb instance for every
// cockroach instance, with each spicedb configured to talk to the corresponding
// cockraoch node.
func NewClusterFromCockroachCluster(c cockroach.Cluster, opts ...SpiceDbOption) Cluster {
	ss := make([]*Node, 0, len(c))

	// the prototypical node will be used to generate a set of nodes
	proto := NewSpiceDbWithOptions(WithTestDefaults(opts...))

	for i := 0; i < len(c); i++ {
		ss = append(ss, &Node{
			ID:            strconv.Itoa(i + 1),
			PresharedKey:  proto.PresharedKey,
			Datastore:     "cockroachdb",
			URI:           c[i].ConnectionString(proto.DbName),
			GrpcPort:      proto.GrpcPort + 2*i,
			InternalPort:  proto.InternalPort + 2*i,
			MetricsPort:   proto.MetricsPort + i,
			DashboardPort: proto.DashboardPort + i,
		})
	}
	return ss
}

// Start starts the entire cluster of spicedb instances
func (c *Cluster) Start(ctx context.Context, out io.Writer, prefix string, args ...string) error {
	for _, s := range *c {
		fmt.Fprintln(out, "starting spice node", s.ID)
		if err := s.Start(ctx, prefix, args...); err != nil {
			return err
		}
	}
	return nil
}

// Stop stops the entire cluster of spicedb instances
func (c *Cluster) Stop(out io.Writer) error {
	for _, s := range *c {
		fmt.Fprintln(out, "stopping spice node", s.ID)
		if err := s.Stop(); err != nil {
			return err
		}
	}
	return nil
}

// Connect blocks until a connection can be made to each instance in the cluster
func (c *Cluster) Connect(ctx context.Context, out io.Writer) error {
	for _, s := range *c {
		fmt.Fprintln(out, "connecting to", s.GrpcPort)
		if err := s.Connect(ctx, out); err != nil {
			return err
		}
	}
	return nil
}

// MigrateHead migrates a Datastore to the latest revision defined in spicedb
func MigrateHead(ctx context.Context, datastore, uri string) error {
	for i := 0; i < 5; i++ {
		if err := e2e.Run(ctx, os.Stdout, os.Stderr,
			"./spicedb",
			"migrate", "head", "--datastore-engine="+datastore,
			"--datastore-conn-uri="+uri,
		); err == nil {
			return nil
		}
		time.Sleep(1 * time.Second)
	}
	return fmt.Errorf("failed to migrate spicedb")
}