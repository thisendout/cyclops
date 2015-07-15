package main

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/fsouza/go-dockerclient"
)

type DockerService interface {
	AttachToContainer(docker.AttachToContainerOptions) error
	CommitContainer(docker.CommitContainerOptions) (*docker.Image, error)
	ContainerChanges(string) ([]docker.Change, error)
	CreateContainer(docker.CreateContainerOptions) (*docker.Container, error)
	RemoveContainer(docker.RemoveContainerOptions) error
	StartContainer(string, *docker.HostConfig) error
	WaitContainer(string) (int, error)
}

func NewDockerClient(host string, tlsVerify string, certPath string) (client *docker.Client, err error) {
	if host == "" {
		return nil, errors.New("DOCKER_HOST must be set")
	}

	if tlsVerify == "yes" || tlsVerify == "1" {
		if certPath == "" {
			return client, errors.New("DOCKER_TLS_VERIFY set without DOCKER_CERT_PATH")
		}

		cert := path.Join(certPath, "cert.pem")
		key := path.Join(certPath, "key.pem")
		ca := path.Join(certPath, "ca.pem")

		client, err = docker.NewTLSClient(host, cert, key, ca)
		if err != nil {
			return client, err
		}
	} else {
		client, err = docker.NewClient(host)
		if err != nil {
			return client, err
		}
	}

	return client, nil
}

func Eval(d DockerService, command string, image string) (EvalResult, error) {
	res := EvalResult{
		Command: command,
		Image:   image,
	}

	cwd, _ := os.Getwd()

	options := docker.CreateContainerOptions{
		Config: &docker.Config{
			Image: image,
			Cmd:   []string{"/bin/bash", "-c", command},
		},
		HostConfig: &docker.HostConfig{
			Binds: []string{fmt.Sprintf("%s:/work", cwd)},
		},
	}
	cont, err := d.CreateContainer(options)
	if err != nil {
		return res, err
	}

	buf := NewBuffer(os.Stdout)
	attachOpts := docker.AttachToContainerOptions{
		Container:    cont.ID,
		OutputStream: buf,
		ErrorStream:  buf,
		Logs:         true,
		Stream:       true,
		Stdin:        false,
		Stdout:       true,
		Stderr:       true,
	}

	go func() {
		d.AttachToContainer(attachOpts)
	}()

	if err := d.StartContainer(cont.ID, &docker.HostConfig{}); err != nil {
		return res, err
	}

	res.Code, err = d.WaitContainer(cont.ID)
	if err != nil {
		return res, err
	}
	res.Log = buf

	if res.Code == 0 {
		if image, err := d.CommitContainer(docker.CommitContainerOptions{Container: cont.ID}); err == nil {
			res.NewImage = image.ID
		}
	}

	res.Changes, err = d.ContainerChanges(cont.ID)
	if err != nil {
		return res, err
	}

	if err := d.RemoveContainer(docker.RemoveContainerOptions{ID: cont.ID}); err != nil {
		return res, err
	}
	return res, nil
}
