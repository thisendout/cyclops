package main

import (
	"bytes"
	"fmt"
	"os"
	"path"

	"github.com/fatih/color"
	"github.com/fsouza/go-dockerclient"
)

func eval(command string, image string) error {
	certPath := os.Getenv("DOCKER_CERT_PATH")
	host := os.Getenv("DOCKER_HOST")

	cert := path.Join(certPath, "cert.pem")
	key := path.Join(certPath, "key.pem")
	ca := path.Join(certPath, "ca.pem")

	client, err := docker.NewTLSClient(host, cert, key, ca)
	if err != nil {
		return err
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
	cont, err := client.CreateContainer(options)
	if err != nil {
		return err
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
		client.AttachToContainer(attachOpts)
	}()

	if err := client.StartContainer(cont.ID, &docker.HostConfig{}); err != nil {
		return err
	}

	_, err = client.WaitContainer(cont.ID)
	if err != nil {
		return err
	}

	changes, err := client.ContainerChanges(cont.ID)
	if err != nil {
		return err
	}
	for _, change := range changes {
		if change.Path == "/work" {
			continue
		}
		switch change.Kind {
		case 0:
			color.Yellow("~ %s", change.Path)
		case 1:
			color.Green("+ %s", change.Path)
		case 2:
			color.Red("- %s", change.Path)
		}
	}

	fmt.Println(string(buf.Bytes()))

	if err := client.RemoveContainer(docker.RemoveContainerOptions{ID: cont.ID}); err != nil {
		return err
	}
	return nil
}
