package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDockerClientSocket(t *testing.T) {
	assert := assert.New(t)

	_, err := NewDockerClient("unix://var/run/docker.sock", "", "")
	assert.NoError(err)
}

func TestNewDockerClientTCPSecure(t *testing.T) {
	assert := assert.New(t)

	_, err := NewDockerClient("tcp://192.168.254.254:2376/", "yes", "")
	assert.EqualError(err, "DOCKER_TLS_VERIFY set without DOCKER_CERT_PATH")

	_, err = NewDockerClient("tcp://192.168.254.254:2376/", "1", "")
	assert.EqualError(err, "DOCKER_TLS_VERIFY set without DOCKER_CERT_PATH")

	_, err = NewDockerClient("tcp://192.168.254.254:2376/", "1", "fixtures/certs/")
	assert.Error(err)
}

func TestNewDockerClientTCPInsecure(t *testing.T) {
	assert := assert.New(t)

	_, err := NewDockerClient("tcp://192.168.254.254:2376/", "no", "")
	assert.NoError(err)

	_, err = NewDockerClient("tcp://192.168.254.254:2376/", "", "")
	assert.NoError(err)
}
