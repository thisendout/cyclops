package main

import (
	"fmt"
	"io"
	"os"
	"strings"

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

func ansible(playbook string, image string) {
	if err := eval("ansible-playbook -i /work/inventory /work/"+playbook, image); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Ansible completed")
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

	line := liner.NewLiner()
	defer line.Close()

	prompt := defaultPrompt
	image := defaultImage

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
					if err := dc.Eval(cmd, image); err != nil {
						fmt.Println(err)
					}
				} else if strings.HasPrefix(input, ":from") {
					line.AppendHistory(input)
					image = strings.TrimPrefix(input, ":from ")
					fmt.Println("Image: ", image)
				} else if strings.HasPrefix(input, ":ansible") {
					line.AppendHistory(input)
					playbook := strings.TrimPrefix(input, ":ansible ")
					ansible(playbook, image)
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
