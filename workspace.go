package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"code.google.com/p/go-uuid/uuid"
	"github.com/fsouza/go-dockerclient"
)

type EvalResult struct {
	Build     bool
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
	CurrentImage string
	history      []EvalResult
	docker       DockerService
}

func NewWorkspace(docker DockerService, mode string, image string) *Workspace {
	ws := &Workspace{
		Mode:         mode,
		Image:        image,
		CurrentImage: image,
		history:      []EvalResult{},
		docker:       docker,
	}
	return ws
}

func (w *Workspace) SetImage(image string) error {
	if err := verifyImage(w.docker, image); err != nil {
		return err
	}
	if w.CurrentImage == w.Image {
		w.CurrentImage = image
	}
	w.Image = image
	return nil
}

func (w *Workspace) CommitLast() (string, error) {
	history := w.history
	if len(history) == 0 {
		return "", errors.New("No container found to commit")
	}
	lastResult := len(w.history) - 1
	if history[lastResult].NewImage != "" && history[lastResult].Id != "" {
		return "", errors.New("Container already committed")
	}

	image := history[lastResult].NewImage
	if history[lastResult].Id != "" {
		committedImage, err := w.commit(history[lastResult].Id)
		if err != nil {
			return "", err
		}
		image = committedImage
		history[lastResult].NewImage = committedImage
	}
	history[lastResult].Deleted = false
	w.history = history
	return image, nil
}

// Run runs Eval but also auto-commits on return code 0
func (w *Workspace) Run(command string) (EvalResult, error) {
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
func (w *Workspace) Eval(command string) (EvalResult, error) {
	res, err := w.evalCommand(command)
	res.Deleted = true
	w.history = append(w.history, res)
	return res, err
}

func (w *Workspace) evalCommand(command string) (EvalResult, error) {
	res, err := Eval(w.docker, command, w.CurrentImage)
	res.BaseImage = w.Image
	return res, err
}

func (w *Workspace) BuildCommit(cmd string) (EvalResult, error) {
	res, err := w.build(cmd)
	w.CurrentImage = res.NewImage
	w.history = append(w.history, res)
	return res, err
}

func (w *Workspace) Build(cmd string) (EvalResult, error) {
	res, err := w.build(cmd)
	res.Deleted = true
	w.history = append(w.history, res)
	return res, err
}

func (w *Workspace) build(cmd string) (EvalResult, error) {
	name := uuid.New()
	c := fmt.Sprintf("FROM %s\n%s", w.CurrentImage, cmd)
	f, _ := os.Create("." + name)
	f.Write([]byte(c))
	res, err := DockerBuild(w.docker, name)
	os.Remove("." + name)
	if err != nil {
		fmt.Println(err)
	}
	res.Build = true
	res.Command = cmd
	res.BaseImage = w.Image
	res.Image = w.CurrentImage
	return res, err
}

type ResetResult struct {
	Err error
	Id  string
}

// Reset cleans up all containers in the history
func (w *Workspace) Reset() (results []ResetResult) {
	history := w.history
	for i, entry := range history {
		if !history[i].Deleted {
			err := RemoveContainer(w.docker, entry.Id)
			results = append(results, ResetResult{Err: err, Id: entry.Id})
			history[i].Deleted = true
		}
	}
	w.history = history
	w.CurrentImage = w.Image
	return
}

func (w *Workspace) Sprint() ([]string, error) {
	res := []string{"FROM " + w.Image}
	for _, entry := range w.history {
		if !entry.Deleted {
			res = append(res, "RUN "+entry.Command)
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
		w.CurrentImage = imageId
	}
	return imageId, err
}

func (w *Workspace) back(n int) error {
	history := w.history
	deleted := 0
	if n > len(history) {
		return errors.New("no history that far back")
	}
	for i := len(history) - 1; i > -1; i -= 1 {
		if history[i].Deleted {
			continue
		}
		RemoveContainer(w.docker, history[i].Id)
		history[i].Deleted = true
		deleted += 1
		if deleted == n {
			w.CurrentImage = history[i].Image
			break
		}
	}
	w.history = history
	return nil
}
