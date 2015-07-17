package main

import (
	"testing"

	"github.com/fsouza/go-dockerclient"
	"github.com/stretchr/testify/assert"
)

func TestParseCommand(t *testing.T) {
	assert := assert.New(t)

	cases := []struct {
		Input   string
		Command string
		Args    string
		Error   error
	}{
		{":help me", "help", "", nil},
		{":help", "help", "", nil},
		{":h", "help", "", nil},
		{":quit", "quit", "", nil},
		{":q", "quit", "", nil},
		{":commit this", "commit", "", nil},
		{":commit", "commit", "", nil},
		{":c", "commit", "", nil},
		{":print this", "print", "", nil},
		{":print", "print", "", nil},
		{":p", "print", "", nil},
		{":eval apt-get update", "eval", "apt-get update", nil},
		{":e apt-get update", "eval", "apt-get update", nil},
		{":eval", "eval", "", ErrMissingRequiredArg},
		{":e", "eval", "", ErrMissingRequiredArg},
		{"apt-get update", "eval", "apt-get update", nil},
		{":run apt-get update", "run", "apt-get update", nil},
		{":r apt-get update", "run", "apt-get update", nil},
		{":run", "run", "", ErrMissingRequiredArg},
		{":r", "run", "", ErrMissingRequiredArg},
		{":from ubuntu:latest", "from", "ubuntu:latest", nil},
		{":f ubuntu:latest", "from", "ubuntu:latest", nil},
		{":from", "from", "", ErrMissingRequiredArg},
		{":f", "from", "", ErrMissingRequiredArg},
		{":write Dockerfile", "write", "Dockerfile", nil},
		{":w Dockerfile", "write", "Dockerfile", nil},
		{":write", "write", "", ErrMissingRequiredArg},
		{":w", "write", "", ErrMissingRequiredArg},
		{":notreal", ":notreal", "", ErrInvalidCommand},
		{"", "", "", nil},
	}
	for _, line := range cases {
		cmd, args, err := parseCommand(line.Input)
		assert.Equal(line.Command, cmd, "Command should be equal for %s", line.Input)
		assert.Equal(line.Args, args, "Args should be equal for %s", line.Input)
		if line.Error == nil {
			assert.NoError(err)
		} else {
			assert.EqualError(err, line.Error.Error())
		}
	}
}

func TestPruneChanges(t *testing.T) {
	assert := assert.New(t)

	changes := []docker.Change{
		{Path: "/tmp", Kind: 0},
		{Path: "/tmp/foo", Kind: 1},
		{Path: "/tmp/foo/banana", Kind: 1},
		{Path: "/tmp/bar", Kind: 2},
		{Path: "/tmp/bar/banana", Kind: 2},
	}

	result := pruneChanges(changes)
	expected := []docker.Change{
		{Path: "/tmp/foo", Kind: 1},
		{Path: "/tmp/foo/banana", Kind: 1},
		{Path: "/tmp/bar", Kind: 2},
		{Path: "/tmp/bar/banana", Kind: 2},
	}
	assert.Equal(result, expected, "they should be equal")
}
