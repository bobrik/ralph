# Ralph

Ralph is a service discovery for twemproxy on mesos. It allows you to run
elastic clusters of memcached and redis and change cluster sizes on
the fly without service interruption and with minimum cache loss.

## Running

Ralph uses [marathoner](https://github.com/bobrik/marathoner) to update
twemproxy configuration. You have to run marathoner updaters to use ralph.

Ralph is distributed as docker container, to run it:

```
docker run -d --net host --name ralph bobrik/ralph \
    -u marathoner-updater1:7676,marathoner-updater2:7676 -b 127.0.0.1 \
    -a '-a 0.0.0.0 -s 22222 -i 2000'
```

This would run ralph with two updaters. Ralph would publish
twemproxy pools on `127.0.0.1`. Twemproxy would publish stats
on `0.0.0.0:22222`.

Ralph is pretty lightweight to run it on every node.

## Adding pools

Your pools are managed by [marathon](https://mesosphere.github.io/marathon/),
here is an example of a pool:


```json
{
  "app": {
    "id": "/twemproxy/one",
    "cmd": "exec redis-server --port $PORT",
    "instances": 2,
    "cpus": 0.1,
    "mem": 128,
    "ports": [
      12334
    ],
    "container": {
      "type": "DOCKER",
      "volumes": [],
      "docker": {
        "image": "redis:2.8"
      }
    },
    "healthChecks": [
      {
        "path": "/",
        "protocol": "TCP",
        "portIndex": 0,
        "gracePeriodSeconds": 15,
        "intervalSeconds": 2,
        "timeoutSeconds": 5,
        "maxConsecutiveFailures": 3
      }
    ],
    "labels": {
      "twemproxy_pool": "redis_one",
      "twemproxy_distribution": "ketama",
      "twemproxy_hash": "fnv1a_64",
      "twemproxy_server_failure_limit": "10",
      "twemproxy_hash_tag": "{}",
      "twemproxy_timeout": "500",
      "twemproxy_server_retry_timeout": "8000",
      "twemproxy_redis": "true",
      "twemproxy_auto_eject_hosts": "true"
    }
  }
}
```

To expose your application with a bunch of redis or memcached servers as
twemproxy pool, you need to use application labels. To set configuration
value for your pool, just add prefix `twemproxy_` to the config key
from [twemproxy configuration](https://github.com/twitter/twemproxy#configuration)
and set it as application label. You should also set `twemproxy_pool`,
otherwise your pool will be ignored.
