package main

import (
	"errors"
	"fmt"
	"os"
	"path"
	"time"

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
	InspectImage(string) (*docker.Image, error)
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
	res.Id = cont.ID

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

	start := time.Now()
	if err := d.StartContainer(cont.ID, &docker.HostConfig{}); err != nil {
		return res, err
	}

	res.Code, err = d.WaitContainer(cont.ID)
	if err != nil {
		return res, err
	}
	res.Log = buf
	res.Duration = time.Since(start)

	if res.Code == 0 {
	}

	res.Changes, err = d.ContainerChanges(cont.ID)
	if err != nil {
		return res, err
	}
	return res, nil
}

func CommitContainer(d DockerService, id string) (string, error) {
	if image, err := d.CommitContainer(docker.CommitContainerOptions{Container: id}); err != nil {
		return "", err
	} else {
		return image.ID, nil
	}
}

func RemoveContainer(d DockerService, id string) error {
	return d.RemoveContainer(docker.RemoveContainerOptions{ID: id})
}

func verifyImage(d DockerService, image string) error {
	_, err := d.InspectImage(image)
	return err
}
