package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/fsouza/go-dockerclient"
)

type EvalResult struct {
	Command   string
	Code      int
	Duration  time.Duration
	Log       *Buffer
	Changes   []docker.Change
	Id        string //container ID
	BaseImage string //assumed base image, used during :from switches
	Image     string //image run against
	NewImage  string //image with committed changes
}

type Workspace struct {
	Mode         string
	Image        string //configured base image
	currentImage string
	state        []string
	history      []*EvalResult
	docker       DockerService
}

func NewWorkspace(docker DockerService, mode string, image string) *Workspace {
	ws := &Workspace{
		Mode:         mode,
		Image:        image,
		currentImage: image,
		state:        []string{},
		history:      []*EvalResult{},
		docker:       docker,
	}
	return ws
}

func (w *Workspace) SetImage(image string) error {
	if w.currentImage == w.Image {
		w.currentImage = image
	}
	w.Image = image
	return nil
}

func (w *Workspace) CommitLast() (string, error) {
	if len(w.history) == 0 {
		return "", errors.New("No container found to commit")
	}
	lastResult := w.history[len(w.history)-1]
	if lastResult.NewImage != "" {
		return "", errors.New("Container already committed")
	}

	image, err := w.commit(lastResult.Id)
	if err != nil {
		return "", err
	}
	lastResult.NewImage = image
	w.state = append(w.state, "RUN "+lastResult.Command)
	return image, nil
}

// Run runs Eval but also auto-commits on return code 0
func (w *Workspace) Run(command string) (*EvalResult, error) {
	res, err := w.Eval(command)
	if res.Code == 0 {
		if imageId, err := w.commit(res.Id); err == nil {
			res.NewImage = imageId
		} else {
			fmt.Println(err)
		}
		w.state = append(w.state, "RUN "+command)
	}
	return res, err
}

// Eval runs the command and updates lastContainer
func (w *Workspace) Eval(command string) (*EvalResult, error) {
	res, err := Eval(w.docker, command, w.currentImage)
	res.BaseImage = w.Image
	w.history = append(w.history, &res)
	return &res, err
}

func (w *Workspace) Sprint() ([]string, error) {
	res := []string{"FROM " + w.Image}
	res = append(res, w.state...)
	return res, nil
}

func (w *Workspace) commit(id string) (string, error) {
	imageId, err := CommitContainer(w.docker, id)
	if err == nil {
		w.currentImage = imageId
	}
	return imageId, err
}
