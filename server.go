package main

import (
	"bytes"

	"github.com/fsouza/go-dockerclient"
)

type EvalRequest struct {
	Command string
	Image   string
}

type EvalResult struct {
	Code    int
	Log     *bytes.Buffer
	Changes []docker.Change
}

// Server manages the cumulative state of the REPL
type Server struct {
	docker *DockerClient
	mode   string //[dockerfile, bash, ansible, puppet]
}

func NewServer(dc *DockerClient, mode string) *Server {
	return &Server{
		docker: dc,
		mode:   mode,
	}
}

func (s *Server) Eval(cmd string, image string) (EvalResult, error) {
	res, err := s.docker.Run(cmd, image)
	return res, err
}
