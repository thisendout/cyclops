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

func printResults(res *EvalResult) {
	fmt.Println()
	fmt.Println("Exit:", res.Code)
	fmt.Println("Took:", res.Duration)
	fmt.Println("From:", res.Image)
	if res.NewImage != "" {
		fmt.Println("Committed:", res.NewImage[:12])
	}
	printChanges(res.Changes)
}

func printChanges(changes []docker.Change) {
	fmt.Println("Changes:")
	if len(changes) == 0 {
		fmt.Println("<none>")
	}
	prunedChanges := pruneChanges(changes)
	for _, change := range prunedChanges {
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

func pruneChanges(changes []docker.Change) []docker.Change {
	var c = []docker.Change{}
	var p string
	for i := len(changes)-1; i > 0; i -= 1 {
		if changes[i].Kind == 0 {
			if !strings.Contains(p, changes[i].Path) {
				c = append(c, changes[i])
			}
		} else {
			c = append(c, changes[i])
		}
		p = changes[i].Path
	}
	// reverse the slice results
	for i := len(c)/2-1; i >= 0; i-- {
		opp := len(c)-1-i
		c[i], c[opp] = c[opp], c[i]
	}
	return c
}

func main() {
	dc, err := NewDockerClient(os.Getenv("DOCKER_HOST"), os.Getenv("DOCKER_TLS_VERIFY"), os.Getenv("DOCKER_CERT_PATH"))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	} else {
		fmt.Println("Connected to docker daemon...")
	}

	ws := NewWorkspace(dc, "bash", defaultImage)

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
					if res, err := ws.Run(cmd); err != nil {
						fmt.Println(err)
					} else {
						printResults(res)
					}
				} else if strings.HasPrefix(input, ":commit") {
					line.AppendHistory(input)
					if id, err := ws.CommitLast(); err != nil {
						fmt.Println(err)
					} else {
						fmt.Println("Committed:", id)
					}
				} else if strings.HasPrefix(input, ":from") {
					line.AppendHistory(input)
					image := strings.TrimPrefix(input, ":from ")
					if err := ws.SetImage(image); err != nil {
						fmt.Println("error setting image:", err)
					} else {
						fmt.Println("Image: ", image)
					}
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
					line.AppendHistory(input)
					if res, err := ws.Eval(input); err != nil {
						fmt.Println(err)
					} else {
						printResults(res)
					}
				}
			}
		}

		if f, err := os.Create("/tmp/.sysrepl_history"); err != nil {
			fmt.Println("error writing history:", err)
		} else {
			line.WriteHistory(f)
			f.Close()
		}
	}
}
