package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewWorkspace(t *testing.T) {
	assert := assert.New(t)
	ws := NewWorkspace(NewMockDockerClient(), "dockerfile", "ubuntu:trusty")

	assert.Equal("dockerfile", ws.Mode)
	assert.Equal("ubuntu:trusty", ws.Image)
	assert.Equal("ubuntu:trusty", ws.CurrentImage)
}

func TestSetImage(t *testing.T) {
	assert := assert.New(t)
	ws := NewWorkspace(NewMockDockerClient(), "dockerfile", "ubuntu:trusty")

	assert.Equal("ubuntu:trusty", ws.Image)
	assert.Equal("ubuntu:trusty", ws.CurrentImage)

	ws.SetImage("fedora")
	assert.Equal("fedora", ws.Image)
	assert.Equal("fedora", ws.CurrentImage)
}

func TestWorkspaceEval(t *testing.T) {
	assert := assert.New(t)
	ws := NewWorkspace(NewMockDockerClient(), "dockerfile", "ubuntu:trusty")

	res, err := ws.Eval("date")
	assert.NoError(err)
	assert.Equal("date", res.Command)
	assert.Equal("ubuntu:trusty", res.Image)
}

func TestWorkspaceSprint(t *testing.T) {
	assert := assert.New(t)
	ws := NewWorkspace(NewMockDockerClient(), "dockerfile", "ubuntu:trusty")

	out, err := ws.Sprint()
	assert.NoError(err)
	assert.Len(out, 1)
	assert.Equal("FROM ubuntu:trusty", out[0])
}

func TestWorkspaceWrite(t *testing.T) {
	assert := assert.New(t)
	ws := NewWorkspace(NewMockDockerClient(), "dockerfile", "ubuntu:trusty")

	os.Mkdir(".test", 0755)
	defer os.RemoveAll(".test")

	ws.SetImage("ubuntu:latest")
	ws.Run("touch /tmp")
	assert.NotPanics(func() {
		err := ws.Write(".test/Dockerfile")
		assert.NoError(err)
	})
}

// Workflow tests
func TestWorkflowRun(t *testing.T) {
	assert := assert.New(t)
	mockdock := NewMockDockerClient()
	ws := NewWorkspace(mockdock, "dockerfile", "ubuntu:trusty")

	for i := 1; i < 4; i++ {
		res, err := ws.Run(fmt.Sprintf("cmd%v", i))
		assert.NoError(err)
		assert.Equal(fmt.Sprintf("cmd%v", i), res.Command)
		assert.Equal(0, res.Code)
		assert.NotNil(res.Duration)
		assert.Equal(fmt.Sprintf("c%v", i), res.Id)
		assert.Equal("ubuntu:trusty", res.BaseImage)
		if i == 1 {
			assert.Equal("ubuntu:trusty", res.Image)
		} else {
			assert.Equal(fmt.Sprintf("i%v", (i-1)), res.Image)
		}
		assert.Equal(fmt.Sprintf("i%v", i), res.NewImage)

		assert.Equal(fmt.Sprintf("i%v", i), ws.CurrentImage)
		assert.False(res.Deleted)
		assert.Len(mockdock.Containers, i)
		assert.Len(mockdock.Images, i)
	}

	results := ws.Reset()
	for _, res := range results {
		assert.NoError(res.Err)
	}
	for _, entry := range ws.history {
		assert.True(entry.Deleted)
	}
}

func TestWorkflowEval(t *testing.T) {
	assert := assert.New(t)
	mockdock := NewMockDockerClient()
	ws := NewWorkspace(mockdock, "dockerfile", "ubuntu:trusty")

	for i := 1; i < 4; i++ {
		res, err := ws.Eval(fmt.Sprintf("cmd%v", i))
		assert.NoError(err)
		assert.Equal(fmt.Sprintf("cmd%v", i), res.Command)
		assert.Equal(0, res.Code)
		assert.NotNil(res.Duration)
		assert.Equal(fmt.Sprintf("c%v", i), res.Id)
		assert.Equal("ubuntu:trusty", res.BaseImage)
		assert.Equal("ubuntu:trusty", res.Image)
		assert.Equal("", res.NewImage)
		assert.True(res.Deleted)

		assert.Equal("ubuntu:trusty", ws.CurrentImage)
		assert.Len(mockdock.Containers, i)
		assert.Len(mockdock.Images, 0)
	}
}

func TestWorkflowEvalCommit(t *testing.T) {
	assert := assert.New(t)
	mockdock := NewMockDockerClient()
	ws := NewWorkspace(mockdock, "dockerfile", "ubuntu:trusty")

	for i := 1; i < 4; i++ {
		res, err := ws.Eval(fmt.Sprintf("cmd%v", i))
		assert.NoError(err)
		assert.Equal(fmt.Sprintf("cmd%v", i), res.Command)
		assert.Equal(0, res.Code)
		assert.NotNil(res.Duration)
		assert.Equal(fmt.Sprintf("c%v", i), res.Id)
		assert.Equal("ubuntu:trusty", res.BaseImage)
		if i == 1 {
			assert.Equal("ubuntu:trusty", res.Image)
		} else {
			assert.Equal(fmt.Sprintf("i%v", (i-1)), res.Image)
		}
		assert.Equal("", res.NewImage)
		assert.Len(mockdock.Containers, i)
		assert.Len(mockdock.Images, i-1)

		image, err := ws.CommitLast()
		assert.NoError(err)
		assert.Equal(fmt.Sprintf("i%v", i), image)
		assert.Len(mockdock.Containers, i)
		assert.Len(mockdock.Images, i)
	}

}

func TestEvalCommand(t *testing.T) {
	assert := assert.New(t)
	ws := NewWorkspace(NewMockDockerClient(), "dockerfile", "ubuntu:trusty")

	for i := 1; i < 4; i++ {
		res, err := ws.evalCommand(fmt.Sprintf("cmd%v", i))
		assert.NoError(err)
		assert.Equal(fmt.Sprintf("cmd%v", i), res.Command)
		assert.Equal(0, res.Code)
		assert.NotNil(res.Duration)
		assert.Equal(fmt.Sprintf("c%v", i), res.Id)
		assert.Equal("ubuntu:trusty", res.BaseImage)
		assert.Equal("ubuntu:trusty", res.Image)
		assert.Equal("", res.NewImage)
		assert.False(res.Deleted)

		assert.Equal("ubuntu:trusty", ws.CurrentImage)
	}
}

func TestSprint(t *testing.T) {
	assert := assert.New(t)
	ws := NewWorkspace(NewMockDockerClient(), "dockerfile", "ubuntu:trusty")

	ws.Eval("cmd1")
	ws.Run("cmd2")
	ws.Run("cmd3")

	state, err := ws.Sprint()
	assert.NoError(err)
	expectedState := []string{"FROM ubuntu:trusty", "RUN cmd2", "RUN cmd3"}
	assert.Equal(expectedState, state)
}

func TestWorkflowBack(t *testing.T) {
	assert := assert.New(t)
	ws := NewWorkspace(NewMockDockerClient(), "dockerfile", "ubuntu:trusty")

	run1, _ := ws.Run("cmd1")
	ws.Run("cmd2")
	ws.Run("cmd3")

	err := ws.back(2)
	assert.NoError(err)
	run4, _ := ws.Run("cmd4")

	state, err := ws.Sprint()
	assert.NoError(err)
	expectedState := []string{"FROM ubuntu:trusty", "RUN cmd1", "RUN cmd4"}

	assert.Equal(run4.Image, run1.NewImage)
	assert.Equal(run4.NewImage, ws.CurrentImage)
	assert.True(ws.history[1].Deleted)
	assert.True(ws.history[2].Deleted)
	assert.Equal(expectedState, state)
}
