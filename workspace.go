package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/fsouza/go-dockerclient"
)

type EvalResult struct {
	Command   string
	Code      int
	Deleted   bool
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
	history      []*EvalResult
	docker       DockerService
}

func NewWorkspace(docker DockerService, mode string, image string) *Workspace {
	ws := &Workspace{
		Mode:         mode,
		Image:        image,
		currentImage: image,
		history:      []*EvalResult{},
		docker:       docker,
	}
	return ws
}

func (w *Workspace) SetImage(image string) error {
	if err := verifyImage(w.docker, image); err != nil {
		return err
	}
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
	lastResult.Deleted = false
	return image, nil
}

// Run runs Eval but also auto-commits on return code 0
func (w *Workspace) Run(command string) (*EvalResult, error) {
	res, err := w.evalCommand(command)
	if res.Code == 0 {
		if imageId, err := w.commit(res.Id); err == nil {
			res.NewImage = imageId
		} else {
			fmt.Println(err)
		}
	}
	w.history = append(w.history, res)
	return res, err
}

// Eval runs the command and updates lastContainer
func (w *Workspace) Eval(command string) (*EvalResult, error) {
	res, err := w.evalCommand(command)
	res.Deleted = true
	w.history = append(w.history, res)
	return res, err
}

func (w *Workspace) evalCommand(command string) (*EvalResult, error) {
	res, err := Eval(w.docker, command, w.currentImage)
	res.BaseImage = w.Image
	return &res, err
}

func (w *Workspace) Sprint() ([]string, error) {
	res := []string{"FROM " + w.Image}
	for _, entry := range w.history {
		if !entry.Deleted {
			res = append(res, "RUN " + entry.Command)
		}
	}
	return res, nil
}

// Write writes the output from Sprint to the provided file
//  The file will be created, if necessary and overwrite the contents
//  if it already exists
func (w *Workspace) Write(path string) error {
	lines, err := w.Sprint()
	if err != nil {
		return err
	}
	var out []byte
	for _, line := range lines {
		out = append(out, []byte(line+"\n")...)
	}
	return ioutil.WriteFile(path, out, 0644)
}

func (w *Workspace) commit(id string) (string, error) {
	imageId, err := CommitContainer(w.docker, id)
	if err == nil {
		w.currentImage = imageId
	}
	return imageId, err
}

func (w *Workspace) back(n int) error {
	var deleted = 0
	if n > len(w.history) {
		return errors.New("no history that far back")
	}
	for i := len(w.history)-1; i > -1; i -= 1 {
		if w.history[i].Deleted {
			continue
		}
		w.history[i].Deleted = true
		deleted += 1
		if deleted == n {
			if w.Image == w.history[i].Image {
				w.currentImage = w.Image
			} else {
				w.currentImage = w.history[i-1].NewImage
			}
			break
		}
	}
	return nil
}
