package main

import (
	"testing"

	"github.com/fsouza/go-dockerclient"
	"github.com/stretchr/testify/assert"
)

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
