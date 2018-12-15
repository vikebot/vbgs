# Deployments

## Pre-Requirements

### Linux

Due to a requirement from Elasticsearch run the following in your terminal:

```bash
$ sudo sysctl -w vm.max_map_count=262144
```

Reference: https://github.com/docker-library/elasticsearch/issues/111#issuecomment-268511769

## Building the containers

### Fluentd

A custom docker container is built using the official fluentd image as basis. Additionally we install the `fluent-plugin-elasticsearch` to ship our logs to elasticsearch for storing and indexing operations.

- https://hub.docker.com/r/fluent/fluentd/
- https://github.com/uken/fluent-plugin-elasticsearch#time_key_format
