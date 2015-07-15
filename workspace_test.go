package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewWorkspace(t *testing.T) {
	assert := assert.New(t)
	ws := NewWorkspace(&MockDockerClient{}, "dockerfile", "ubuntu:trusty")

	assert.Equal("dockerfile", ws.Mode)
	assert.Equal("ubuntu:trusty", ws.Image)
	assert.Equal("ubuntu:trusty", ws.currentImage)
}

func TestSetImage(t *testing.T) {
	assert := assert.New(t)
	ws := NewWorkspace(&MockDockerClient{}, "dockerfile", "ubuntu:trusty")

	assert.Equal("ubuntu:trusty", ws.Image)
	assert.Equal("ubuntu:trusty", ws.currentImage)

	ws.SetImage("fedora")
	assert.Equal("fedora", ws.Image)
	assert.Equal("ubuntu:trusty", ws.currentImage)
}

func TestWorkspaceEval(t *testing.T) {
	assert := assert.New(t)
	ws := NewWorkspace(&MockDockerClient{}, "dockerfile", "ubuntu:trusty")

	res, err := ws.Eval("date")
	assert.NoError(err)
	assert.Equal("date", res.Command)
	assert.Equal("ubuntu:trusty", res.Image)
}

func TestWorkspaceSprint(t *testing.T) {
	assert := assert.New(t)
	ws := NewWorkspace(&MockDockerClient{}, "dockerfile", "ubuntu:trusty")

	out, err := ws.Sprint()
	assert.NoError(err)
	assert.Len(out, 1)
	assert.Equal("FROM ubuntu:trusty", out[0])
}
