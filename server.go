package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/fsouza/go-dockerclient"
)

type Server struct {
	docker *docker.Client
}

func NewServer(dc *docker.Client) *Server {
	return &Server{
		docker: dc,
	}
}

func (s *Server) Eval(command string, image string) (EvalResult, error) {
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
	cont, err := s.docker.CreateContainer(options)
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
		s.docker.AttachToContainer(attachOpts)
	}()

	if err := s.docker.StartContainer(cont.ID, &docker.HostConfig{}); err != nil {
		return res, err
	}

	res.Code, err = s.docker.WaitContainer(cont.ID)
	if err != nil {
		return res, err
	}
	res.Log = &buf

	if res.Code == 0 {
		if image, err := s.docker.CommitContainer(docker.CommitContainerOptions{Container: cont.ID}); err == nil {
			res.NewImage = image.ID
		}
	}

	res.Changes, err = s.docker.ContainerChanges(cont.ID)
	if err != nil {
		return res, err
	}

	if err := s.docker.RemoveContainer(docker.RemoveContainerOptions{ID: cont.ID}); err != nil {
		return res, err
	}
	return res, nil
}
