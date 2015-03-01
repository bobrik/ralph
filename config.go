package ralph

import (
	"fmt"
	"net"
	"strconv"

	"github.com/bobrik/marathoner"
)

// twemproxyConfig is a bunch of twemproxy pools
type twemproxyConfig map[string]twemproxyPool

// twemproxyPool represents configuration for twemproxyPool
type twemproxyPool struct {
	Listen             string            `yaml:"listen"`
	Hash               string            `yaml:"hash"`
	HashTag            string            `yaml:"hash_tag,omitempty"`
	Distribution       string            `yaml:"distribution"`
	Timeout            int               `yaml:"timeout,omitempty"`
	Backlog            int               `yaml:"backlog,omitempty"`
	Preconnect         bool              `yaml:"preconnect,omitempty"`
	Redis              bool              `yaml:"redis,omitempty"`
	ServerConnections  int               `yaml:"server_connections,omitempty"`
	AutoEjectHosts     bool              `yaml:"auto_eject_hosts"`
	ServerRetryTimeout int               `yaml:"server_retry_timeout,omitempty"`
	ServerFailureLimit int               `yaml:"server_failure_limit,omitempty"`
	Servers            []twemproxyServer `yaml:"servers"`
}

// twemproxyServer represents server in twemproxy config
type twemproxyServer struct {
	Host   string
	Port   int
	Weight int
	Name   string
}

func (s twemproxyServer) MarshalYAML() (interface{}, error) {
	return fmt.Sprintf("%s:%d:%d %s", s.Host, s.Port, s.Weight, s.Name), nil
}

// newPoolFromApp creates new twemproxy pool from marathoner application
func newPoolFromApp(app marathoner.App, bind string) (twemproxyPool, error) {
	pool := twemproxyPool{
		Servers: []twemproxyServer{},
	}

	if len(app.Ports) == 0 {
		return pool, fmt.Errorf("app %s has no ports assigned", app.Name)
	}

	pool.Listen = net.JoinHostPort(bind, strconv.Itoa(app.Ports[0]))

	intFields := map[string]*int{
		"twemproxy_timeout":              &pool.Timeout,
		"twemproxy_backlog":              &pool.Backlog,
		"twemproxy_server_connections":   &pool.ServerConnections,
		"twemproxy_server_retry_timeout": &pool.ServerRetryTimeout,
		"twemproxy_server_failure_limit": &pool.ServerFailureLimit,
	}

	for f, d := range intFields {
		err := fetchInt(app.Labels, f, d)
		if err != nil {
			return pool, err
		}
	}

	stringFields := map[string]*string{
		"twemproxy_hash":         &pool.Hash,
		"twemproxy_hash_tag":     &pool.HashTag,
		"twemproxy_distribution": &pool.Distribution,
	}

	for f, d := range stringFields {
		fetchString(app.Labels, f, d)
	}

	boolFields := map[string]*bool{
		"twemproxy_preconnect":       &pool.Preconnect,
		"twemproxy_redis":            &pool.Redis,
		"twemproxy_auto_eject_hosts": &pool.AutoEjectHosts,
	}

	for f, d := range boolFields {
		fetchBool(app.Labels, f, d)
	}

	for _, t := range app.Tasks {
		if len(t.Ports) == 0 {
			return pool, fmt.Errorf("task %s of app %s has zero ports", t.ID, app.Name)
		}

		server := twemproxyServer{
			Name:   t.ID,
			Host:   t.Host,
			Port:   t.Ports[0],
			Weight: 1,
		}

		pool.Servers = append(pool.Servers, server)
	}

	return pool, nil
}

// fetchInt fetches int from string field of a map if possible
func fetchInt(source map[string]string, field string, destination *int) error {
	if v, ok := source[field]; ok {
		f, err := strconv.Atoi(v)
		if err != nil {
			return err
		}

		*destination = f
	}

	return nil
}

// fetchString fetches string value from a map if it exists
func fetchString(source map[string]string, field string, destination *string) {
	if v, ok := source[field]; ok {
		*destination = v
	}
}

// fetchBool fetches bool from string field of a map if possible
func fetchBool(source map[string]string, field string, destination *bool) {
	if v, ok := source[field]; ok && (v == "true" || v == "1") {
		*destination = true
	}
}
