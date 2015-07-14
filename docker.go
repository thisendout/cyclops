package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/fsouza/go-dockerclient"
)

type DockerClient struct {
	client *docker.Client
}

func NewDockerClient() (*DockerClient, error) {
	dc := &DockerClient{}

	host := os.Getenv("DOCKER_HOST")
	if host == "" {
		return dc, errors.New("DOCKET_HOST must be set")
	}

	if tlsVerify := os.Getenv("DOCKER_TLS_VERIFY"); tlsVerify == "yes" {
		certPath := os.Getenv("DOCKER_CERT_PATH")
		if certPath == "" {
			return dc, errors.New("DOCKER_TLS_VERIFY set without DOCKER_CERT_PATH")
		}

		cert := path.Join(certPath, "cert.pem")
		key := path.Join(certPath, "key.pem")
		ca := path.Join(certPath, "ca.pem")

		if client, err := docker.NewTLSClient(host, cert, key, ca); err != nil {
			return dc, err
		} else {
			dc.client = client
		}
	} else {
		if client, err := docker.NewClient(host); err != nil {
			return dc, err
		} else {
			dc.client = client
		}
	}

	if err := dc.client.Ping(); err != nil {
		return dc, err
	}
	return dc, nil
}

func (d *DockerClient) Run(command string, image string) (EvalResult, error) {
	res := EvalResult{}

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
	cont, err := d.client.CreateContainer(options)
	if err != nil {
		return res, err
	}

	var buf bytes.Buffer
	attachOpts := docker.AttachToContainerOptions{
		Container:    cont.ID,
		OutputStream: &buf,
		ErrorStream:  &buf,
		Logs:         true,
		Stream:       true,
		Stdin:        false,
		Stdout:       true,
		Stderr:       true,
	}

	go func() {
		d.client.AttachToContainer(attachOpts)
	}()

	if err := d.client.StartContainer(cont.ID, &docker.HostConfig{}); err != nil {
		return res, err
	}

	res.Code, err = d.client.WaitContainer(cont.ID)
	if err != nil {
		return res, err
	}
	res.Log = &buf

	res.Changes, err = d.client.ContainerChanges(cont.ID)
	if err != nil {
		return res, err
	}

	if err := d.client.RemoveContainer(docker.RemoveContainerOptions{ID: cont.ID}); err != nil {
		return res, err
	}
	return res, nil
}