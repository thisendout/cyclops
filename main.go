package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/fsouza/go-dockerclient"
	"github.com/peterh/liner"
)

const (
	defaultPrompt = "sysrepl"
	defaultImage  = "ubuntu:trusty"
)

func help() {
	usage := `sysrepl - help
:help	show help
:from	set base image
:run	execute shell command
:quit	quit sysrepl - <ctrl-d>
`
	fmt.Println(usage)
}

func printChanges(changes []docker.Change) {
	for _, change := range changes {
		if change.Path == "/work" {
			continue
		}
		switch change.Kind {
		case 0:
			color.Yellow("~ %s", change.Path)
		case 1:
			color.Green("+ %s", change.Path)
		case 2:
			color.Red("- %s", change.Path)
		}
	}
}

func main() {
	dc, err := NewDockerClient()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	} else {
		fmt.Println("Connected to docker daemon...")
	}

	server := NewServer(dc)
	ws := NewWorkspace(server, "bash", defaultImage)

	line := liner.NewLiner()
	defer line.Close()

	prompt := defaultPrompt

	if f, err := os.Open("/tmp/.sysrepl_history"); err == nil {
		line.ReadHistory(f)
		f.Close()
	}

mainloop:
	for {
		if input, err := line.Prompt(prompt + "> "); err != nil {
			if err == io.EOF {
				fmt.Println() //Returns user to prompt on a new line
				break
			}
		} else {
			switch input {
			case ":help":
				help()
			case ":quit":
				fmt.Println("Exiting...")
				break mainloop
			default:
				if strings.HasPrefix(input, ":run") {
					cmd := strings.TrimPrefix(input, ":run ")
					line.AppendHistory(input)
					if res, err := ws.Eval(cmd); err != nil {
						fmt.Println(err)
					} else {
						fmt.Println("Exit: ", res.Code)
						fmt.Println(res.Log)
						printChanges(res.Changes)
					}
				} else if strings.HasPrefix(input, ":from") {
					line.AppendHistory(input)
					image := strings.TrimPrefix(input, ":from ")
					ws.SetImage(image)
					fmt.Println("Image: ", image)
				} else if strings.HasPrefix(input, ":print") {
					line.AppendHistory(input)
					if out, err := ws.Sprint(); err == nil {
						for _, line := range out {
							fmt.Println(line)
						}
					}
				} else if input == "" {
					continue
				} else {
					fmt.Println(input)
					line.AppendHistory(input)
				}
			}
		}

		if f, err := os.Create("/tmp/.sysrepl_history"); err != nil {
			fmt.Println("error writing history", err)
		} else {
			line.WriteHistory(f)
			f.Close()
		}
	}
}
