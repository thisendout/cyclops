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

func TestVerifyImage(t *testing.T) {
	assert := assert.New(t)

	md := NewMockDockerClient()

	err := verifyImage(md, "ubuntu:trusty")
	assert.NoError(err)

	md.FailInspect = true
	err = verifyImage(md, "ubuntu:trusty")
	assert.Error(err)
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
	FailInspect  bool
	PleaseReturn int
	lastId       int
	Containers   []*docker.Container
	Images       []*docker.Image
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
		FailInspect:  false,
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
	image := &docker.Image{
		ID: strings.Replace(opts.Container, "c", "i", 1),
	}
	m.Images = append(m.Images, image)

	return image, nil
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
	cont := &docker.Container{
		ID: fmt.Sprintf("c%v", m.lastId),
	}
	m.Containers = append(m.Containers, cont)
	return cont, nil
}

func (m *MockDockerClient) RemoveContainer(opts docker.RemoveContainerOptions) error {
	if m.FailRemove {
		return errors.New("MOCK: Failed to remove container")
	}
	var newContainers []*docker.Container
	if len(m.Containers) == 1 {
		m.Containers = []*docker.Container{}
		return nil
	}
	for i, c := range m.Containers {
		if c.ID == opts.ID {
			newContainers = m.Containers[:i]
			if i < len(m.Containers)-1 {
				newContainers = append(newContainers, m.Containers[(i+1):]...)
			}
			break
		}
	}
	m.Containers = newContainers
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

func (m *MockDockerClient) InspectImage(string) (*docker.Image, error) {
	if m.FailInspect {
		return &docker.Image{}, errors.New("MOCK: Failed to find image")
	}
	return &docker.Image{}, nil
}
