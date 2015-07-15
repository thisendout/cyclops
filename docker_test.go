package main

import (
	"testing"

	"github.com/fsouza/go-dockerclient"
	"github.com/stretchr/testify/assert"
)

func TestNewDockerClientNegative(t *testing.T) {
	assert := assert.New(t)

	_, err := NewDockerClient("", "", "")
	assert.EqualError(err, "DOCKER_HOST must be set")
}

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

	_, err = NewDockerClient("tcp://192.168.254.254:2376/", "0", "")
	assert.NoError(err)

	_, err = NewDockerClient("tcp://192.168.254.254:2376/", "", "")
	assert.NoError(err)
}

func TestEval(t *testing.T) {
	assert := assert.New(t)

	res, err := Eval(&MockDockerClient{}, "date", "ubuntu:trusty")
	assert.NoError(err)
	assert.Equal("date", res.Command)
	assert.Equal("ubuntu:trusty", res.Image)
}

//Mock Docker Client for use by Servers and Workspaces for testing
type MockDockerClient struct{}

func (m *MockDockerClient) AttachToContainer(docker.AttachToContainerOptions) error {
	return nil
}

func (m *MockDockerClient) CommitContainer(docker.CommitContainerOptions) (*docker.Image, error) {
	return &docker.Image{}, nil
}

func (m *MockDockerClient) ContainerChanges(string) ([]docker.Change, error) {
	return []docker.Change{}, nil
}

func (m *MockDockerClient) CreateContainer(docker.CreateContainerOptions) (*docker.Container, error) {
	return &docker.Container{}, nil
}

func (m *MockDockerClient) RemoveContainer(docker.RemoveContainerOptions) error {
	return nil
}

func (m *MockDockerClient) StartContainer(string, *docker.HostConfig) error {
	return nil
}

func (m *MockDockerClient) WaitContainer(string) (int, error) {
	return 0, nil
}
