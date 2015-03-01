package ralph

import (
	"reflect"
	"testing"

	"github.com/bobrik/marathoner"
	"gopkg.in/yaml.v2"
)

func TestMarshalling(t *testing.T) {
	c := twemproxyConfig(map[string]twemproxyPool{
		"my_pool": twemproxyPool{
			Listen:       "0.0.0.0:6379",
			Hash:         "fnv1_64",
			Distribution: "ketama",
			Servers: []twemproxyServer{
				{
					Name:   "one",
					Host:   "one.dev",
					Port:   6381,
					Weight: 1,
				},
				{
					Name:   "two",
					Host:   "two.dev",
					Port:   6382,
					Weight: 2,
				},
				{
					Name:   "three",
					Host:   "three.dev",
					Port:   6383,
					Weight: 8,
				},
			},
		},
	})

	r, err := yaml.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}

	expected := (`my_pool:
  listen: 0.0.0.0:6379
  hash: fnv1_64
  distribution: ketama
  auto_eject_hosts: false
  servers:
  - one.dev:6381:1 one
  - two.dev:6382:2 two
  - three.dev:6383:8 three
`)

	if expected != string(r) {
		t.Fatalf("expected:\n%s\ngot:\n%s\n", expected, string(r))
	}
}

func TestAppIncompleteConversion(t *testing.T) {
	app := marathoner.App{
		Name:  "whatever",
		Ports: []int{12345},
		Labels: map[string]string{
			"twemproxy_timeout":          "800",
			"twemproxy_backlog":          "1024",
			"twemproxy_redis":            "true",
			"twemproxy_auto_eject_hosts": "false",
			"twemproxy_hash_tag":         "{}",
		},
	}

	pool, err := newPoolFromApp(app, "192.168.6.6")
	if err != nil {
		t.Fatal(err)
	}

	expected := twemproxyPool{
		Listen:  "192.168.6.6:12345",
		HashTag: "{}",
		Timeout: 800,
		Backlog: 1024,
		Redis:   true,
		Servers: []twemproxyServer{},
	}

	if !reflect.DeepEqual(expected, pool) {
		t.Fatalf("expected %#v, got %#v", expected, pool)
	}
}

func TestFullConversion(t *testing.T) {
	app := marathoner.App{
		Name:  "whatever",
		Ports: []int{12345},
		Labels: map[string]string{
			"twemproxy_hash":                 "hoho",
			"twemproxy_hash_tag":             "{}",
			"twemproxy_distribution":         "some",
			"twemproxy_timeout":              "800",
			"twemproxy_backlog":              "1024",
			"twemproxy_preconnect":           "true",
			"twemproxy_redis":                "true",
			"twemproxy_server_connections":   "8",
			"twemproxy_auto_eject_hosts":     "true",
			"twemproxy_server_retry_timeout": "200",
			"twemproxy_server_failure_limit": "5",
		},
		Tasks: []marathoner.Task{
			{
				ID:    "id1",
				Host:  "server-1",
				Ports: []int{30000},
			},
			{
				ID:    "id2",
				Host:  "server-2",
				Ports: []int{30002},
			},
			{
				ID:    "id3",
				Host:  "server-3",
				Ports: []int{30008},
			},
		},
	}

	pool, err := newPoolFromApp(app, "192.168.6.6")
	if err != nil {
		t.Fatal(err)
	}

	expected := twemproxyPool{
		Listen:             "192.168.6.6:12345",
		Hash:               "hoho",
		HashTag:            "{}",
		Distribution:       "some",
		Timeout:            800,
		Backlog:            1024,
		Preconnect:         true,
		Redis:              true,
		ServerConnections:  8,
		AutoEjectHosts:     true,
		ServerRetryTimeout: 200,
		ServerFailureLimit: 5,
		Servers: []twemproxyServer{
			{
				Name:   "id1",
				Host:   "server-1",
				Port:   30000,
				Weight: 1,
			},
			{
				Name:   "id2",
				Host:   "server-2",
				Port:   30002,
				Weight: 1,
			},
			{
				Name:   "id3",
				Host:   "server-3",
				Port:   30008,
				Weight: 1,
			},
		},
	}

	if !reflect.DeepEqual(expected, pool) {
		t.Fatalf("expected %#v, got %#v", expected, pool)
	}
}
