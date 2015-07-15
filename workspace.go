package main

import (
	"github.com/fsouza/go-dockerclient"
)

type EvalRequest struct {
	Command string
	Image   string
}

type EvalResult struct {
	Code     int
	Log      *Buffer
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
