package main

import (
	"errors"
	"fmt"
	"strings"
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

	res, err := Eval(NewMockDockerClient(), "date", "ubuntu:trusty")
	assert.NoError(err)
	assert.Equal("date", res.Command)
	assert.Equal("ubuntu:trusty", res.Image)
}

//Mock Docker Client for use by Servers and Workspaces for testing
type MockDockerClient struct {
	FailAttach   bool
	FailChanges  bool
	FailCommit   bool
	FailCreate   bool
	FailRemove   bool
	FailStart    bool
	FailWait     bool
	PleaseReturn int
	lastId       int
}

func NewMockDockerClient() *MockDockerClient {
	return &MockDockerClient{
		FailAttach:   false,
		FailChanges:  false,
		FailCommit:   false,
		FailCreate:   false,
		FailRemove:   false,
		FailStart:    false,
		FailWait:     false,
		PleaseReturn: 0,
		lastId:       0,
	}
}

func (m *MockDockerClient) AttachToContainer(docker.AttachToContainerOptions) error {
	if m.FailAttach {
		return errors.New("MOCK: Failed to attach")
	}
	return nil
}

func (m *MockDockerClient) CommitContainer(opts docker.CommitContainerOptions) (*docker.Image, error) {
	if m.FailCommit {
		return &docker.Image{}, errors.New("MOCK: Failed to commit")
	}
	return &docker.Image{
		ID: strings.Replace(opts.Container, "c", "i", 1),
	}, nil
}

func (m *MockDockerClient) ContainerChanges(string) ([]docker.Change, error) {
	if m.FailChanges {
		return []docker.Change{}, errors.New("MOCK: Failed to determine changes")
	}
	return []docker.Change{}, nil
}

func (m *MockDockerClient) CreateContainer(docker.CreateContainerOptions) (*docker.Container, error) {
	if m.FailCreate {
		return &docker.Container{}, errors.New("MOCK: Failed to create container")
	}
	m.lastId++
	return &docker.Container{
		ID: fmt.Sprintf("c%v", m.lastId),
	}, nil
}

func (m *MockDockerClient) RemoveContainer(docker.RemoveContainerOptions) error {
	if m.FailRemove {
		return errors.New("MOCK: Failed to remove container")
	}
	return nil
}

func (m *MockDockerClient) StartContainer(string, *docker.HostConfig) error {
	if m.FailStart {
		return errors.New("MOCK: Failed to start container")
	}
	return nil
}

func (m *MockDockerClient) WaitContainer(string) (int, error) {
	if m.FailWait {
		return m.PleaseReturn, errors.New("MOCK: Failed to wait on container")
	}
	return m.PleaseReturn, nil
}
