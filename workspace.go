package main

import (
	"github.com/fsouza/go-dockerclient"
)

type EvalResult struct {
	Command  string
	Code     int
	Log      *Buffer
	Changes  []docker.Change
	Image    string //image run against
	NewImage string //image with committed changes
}

type Workspace struct {
	Mode         string
	Image        string
	currentImage string
	state        []string
	history      []EvalResult
	docker       DockerService
}

func NewWorkspace(docker DockerService, mode string, image string) *Workspace {
	ws := &Workspace{
		Mode:         mode,
		Image:        image,
		currentImage: image,
		state:        []string{},
		history:      []EvalResult{},
		docker:       docker,
	}
	return ws
}

func (w *Workspace) SetImage(image string) error {
	w.Image = image
	return nil
}

func (w *Workspace) Eval(command string) (EvalResult, error) {
	res, err := Eval(w.docker, command, w.currentImage)
	if res.Code == 0 {
		w.currentImage = res.NewImage
		w.state = append(w.state, "RUN "+command)
	}
	w.history = append(w.history, res)
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
