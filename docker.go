package main

import (
	"errors"
	"path"

	"github.com/fsouza/go-dockerclient"
)

type DockerClient struct {
	*docker.Client
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
