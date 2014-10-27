package elasticache

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"regexp"
	"strconv"
	"time"
)

// defaultTimeout is the default socket read/write timeout.
const defaultTimeout = time.Duration(100) * time.Millisecond

var (
	crlf  = []byte("\r\n")
	lf    = []byte("\n")
	space = []byte(" ")
	pipe  = []byte("|")
)

// ElastiCacheServer represents a single cache server returned by the
// Auto Discover service. You can query its Hostname, IP Address and Port
// number.
type ElastiCacheServer struct {
	Hostname string
	Address  string
	Port     int
}

// AutoDiscoverer is an ElastiCache Auto Discovery client.
type AutoDiscoverer struct {
	// Elasticache Configuration Endpoint DNS Name
	endpoint string
}

// NewAutoDiscover takes an ElastiCache configuration server address string
// like "foo.cfg.amazonaws.com:11211", and returns an *AutoDiscoverer and error.
func NewAutoDiscoverer(configServer string) (*AutoDiscoverer, error) {
	return &AutoDiscoverer{configServer}, nil
}

// GetClusterConfig returns the current ElastiCache cluster configuration:
// the current version number and a slice of ElastiCache server
// hostname/address/port tuples. It is recommended that you call this function
// once every 60 seconds, keep track of the previously seen version number
// and only replace the active servers in the Memcache client when the version
// number increases.
func (a *AutoDiscoverer) GetClusterConfig() (int, []ElastiCacheServer, error) {
	var (
		addr net.Addr
		err  error
	)

	if addr, err = net.ResolveTCPAddr("tcp", a.endpoint); err != nil {
		return 0, nil, fmt.Errorf("Cannot resolve tcp address: %v", err)
	}

	cn, err := connect(addr)
	if err != nil {
		return 0, nil, err
	}
	defer cn.nc.Close()

	if _, err = fmt.Fprintf(cn.rw, "config get cluster\r\n"); err != nil {
		return 0, nil, err
	}
	if err = cn.rw.Flush(); err != nil {
		return 0, nil, err
	}

	// Get the raw form of the server list
	version, servers, err := parseConfigResponse(cn.rw.Reader)
	if err != nil {
		return 0, nil, err
	}

	// Pull out config version and split server string up
	var elastiCacheServers []ElastiCacheServer
	serverList := bytes.Split(servers, space)
	for _, elem := range serverList {
		server := bytes.Split(elem, pipe)

		if len(server) != 3 {
			return 0, nil, fmt.Errorf("elasticache: insufficient elements returned for server %s", server)
		}

		port, _ := strconv.Atoi(string(server[2]))

		elastiCacheServers = append(
			elastiCacheServers,
			ElastiCacheServer{
				string(server[0]),
				string(server[1]),
				port,
			},
		)
	}

	return version, elastiCacheServers, nil
}

// conn is a connection to a server.
type conn struct {
	nc   net.Conn
	rw   *bufio.ReadWriter
	addr net.Addr
}

func (cn *conn) extendDeadline() {
	cn.nc.SetDeadline(time.Now().Add(defaultTimeout))
}

// ConnectTimeoutError is the error type used when it takes
// too long to connect to the desired host. This level of
// detail can generally be ignored.
type ConnectTimeoutError struct {
	Addr net.Addr
}

func (cte *ConnectTimeoutError) Error() string {
	return "elasticache: connect timeout to " + cte.Addr.String()
}

func connect(addr net.Addr) (*conn, error) {
	nc, err := net.DialTimeout(addr.Network(), addr.String(), defaultTimeout)
	if err != nil {
		if ne, ok := err.(net.Error); ok && ne.Timeout() {
			return &conn{}, &ConnectTimeoutError{addr}
		}
		return &conn{}, err
	}

	cn := &conn{
		nc:   nc,
		addr: addr,
		rw:   bufio.NewReadWriter(bufio.NewReader(nc), bufio.NewWriter(nc)),
	}
	cn.extendDeadline()
	return cn, nil
}

// parseConfigResponse reads a CONFIG response from r and returns the version
// number and byte array of server descriptions.
func parseConfigResponse(r *bufio.Reader) (int, []byte, error) {
	line, err := r.ReadSlice('\n')
	if err != nil {
		return 0, nil, err
	}
	size, err := scanConfigDataLength(line)
	if err != nil {
		return 0, nil, err
	}
	value, err := ioutil.ReadAll(io.LimitReader(r, int64(size)+2))
	if err != nil {
		return 0, nil, err
	}
	if !bytes.HasSuffix(value, crlf) {
		return 0, nil, fmt.Errorf("elasticache: corrupt config result read")
	}
	// Reduce the value down to its total identified size
	value = value[:size]

	// Pull out version and server list, separated by LFs
	re := regexp.MustCompile("^(\\d+)\n(.*)\n$")
	configData := re.FindSubmatch(value)
	if configData == nil {
		return 0, nil, fmt.Errorf("elasticache: malformed config data: %s", value)
	}

	version, _ := strconv.Atoi(string(configData[1]))
	return version, configData[2], nil
}

// scanConfigDataLength returns the declared length of the cluster
// configuration. It does not read the bytes of the configuration itself.
func scanConfigDataLength(line []byte) (int, error) {
	pattern := "CONFIG cluster 0 %d\r\n"
	size := 0

	n, err := fmt.Sscanf(string(line), pattern, &size)
	if err != nil || n != 1 {
		return size, fmt.Errorf("elasticache: unexpected line in config response: %q", line)
	}

	return size, nil
}
