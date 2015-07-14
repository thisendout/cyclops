package main

import (
	"errors"
	"os"
	"path"

	"github.com/fsouza/go-dockerclient"
)

func NewDockerClient() (client *docker.Client, err error) {
	host := os.Getenv("DOCKER_HOST")
	if host == "" {
		return nil, errors.New("DOCKET_HOST must be set")
	}

	if tlsVerify := os.Getenv("DOCKER_TLS_VERIFY"); tlsVerify == "yes" {
		certPath := os.Getenv("DOCKER_CERT_PATH")
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

	if err := client.Ping(); err != nil {
		return client, err
	}
	return client, nil
}
