package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/peterh/liner"
	"github.com/rjeczalik/notify"
)

const (
	defaultPrompt = "sysrepl"
)

func help() {
	usage := `sysrepl - help
:bash	execute bash script
:help	show help
:image	set base image
:quit	quit sysrepl - <ctrl-d>
`
	fmt.Println(usage)
}

func handle(ei notify.EventInfo) {
	fmt.Println(ei)

	if err := eval("/work/run.sh", "ubuntu:trusty"); err != nil {
		fmt.Println(err)
	}
}

func ansible(playbook string, image string) {
	if err := eval("ansible-playbook -i /work/inventory /work/"+playbook, image); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Ansible completed")
	}
}

func watch() {
	c := make(chan notify.EventInfo, 1)

	if err := notify.Watch("./...", c, notify.All); err != nil {
		panic(err)
	}
	defer notify.Stop(c)

	for {
		ei := <-c
		handle(ei)
	}
}

func main() {
	line := liner.NewLiner()
	defer line.Close()

	prompt := defaultPrompt
	image := "ubuntu:trusty"

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
				if strings.HasPrefix(input, ":bash") {
					cmd := strings.TrimPrefix(input, ":bash ")
					line.AppendHistory(input)
					if err := eval(cmd, image); err != nil {
						fmt.Println(err)
					}
				} else if strings.HasPrefix(input, ":image") {
					image = strings.TrimPrefix(input, ":image ")
					fmt.Println("Image: ", image)
				} else if strings.HasPrefix(input, ":ansible") {
					playbook := strings.TrimPrefix(input, ":ansible ")
					ansible(playbook, image)
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
