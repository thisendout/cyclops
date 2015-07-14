package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/fsouza/go-dockerclient"
)

type EvalRequest struct {
	Command string
	Image   string
}

type EvalResult struct {
	Code     int
	Log      *bytes.Buffer
	Changes  []docker.Change
	NewImage string
}

type Workspace struct {
	Mode         string
	Image        string
	currentImage string
	state        []string
	server       *Server
}

func NewWorkspace(server *Server, mode string, image string) *Workspace {
	ws := &Workspace{
		Mode:         mode,
		Image:        image,
		currentImage: image,
		state:        []string{},
		server:       server,
	}
	return ws
}

func (w *Workspace) SetImage(image string) error {
	w.Image = image
	return nil
}

func (w *Workspace) Eval(command string) (EvalResult, error) {
	res, err := w.server.Eval(command, w.currentImage)
	if res.Code == 0 {
		w.currentImage = res.NewImage
		w.state = append(w.state, "RUN "+command)
	}
	return res, err
}

func (w *Workspace) Sprint() ([]string, error) {
	res := []string{"FROM " + w.Image}
	res = append(res, w.state...)
	return res, nil
}

func (w *Workspace) Write(outfile string) error {
	return nil
}

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
