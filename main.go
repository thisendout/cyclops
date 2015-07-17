package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"strconv"

	"github.com/fatih/color"
	"github.com/fsouza/go-dockerclient"
	"github.com/peterh/liner"
)

const (
	defaultPrompt = "cyclops"
	defaultImage  = "ubuntu:trusty"
)

var (
	ErrInvalidCommand     = errors.New("Invalid command")
	ErrMissingRequiredArg = errors.New("Missing required argument for command")
)

func help() {
	usage := `cyclops - help
:help|:h                     show help
:from|:f      [image]        set base image
:eval|:e      [command ...]  execute shell command (ephemeral)
:run|:r       [command ...]  execute shell command (auto commits image)
:commit|:c                   commit changes from last command
:back|:b      [num]          go back in the history (default: 1)
:history|:hs                 show the current history
:write|:w     [path/to/file] write state to file
:quit|:q                     quit cyclops - <ctrl-d>
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
	for i := len(changes) - 1; i > 0; i -= 1 {
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
	for i := len(c)/2 - 1; i >= 0; i-- {
		opp := len(c) - 1 - i
		c[i], c[opp] = c[opp], c[i]
	}
	return c
}

func parseCommand(input string) (string, string, error) {
	if input == "" {
		return "", "", nil
	}
	if string(input[0]) != ":" {
		return "eval", input, nil
	}
	parts := strings.SplitN(input, " ", 2)
	switch parts[0] {
	case ":commit", ":c":
		return "commit", "", nil
	case ":help", ":h":
		return "help", "", nil
	case ":print", ":p":
		return "print", "", nil
	case ":history", ":hs":
		return "history", "", nil
	case ":quit", ":q":
		return "quit", "", nil
	case ":eval", ":e":
		if len(parts) < 2 {
			return "eval", "", ErrMissingRequiredArg
		}
		return "eval", parts[1], nil
	case ":from", ":f":
		if len(parts) < 2 {
			return "from", "", ErrMissingRequiredArg
		}
		return "from", parts[1], nil
	case ":run", ":r":
		if len(parts) < 2 {
			return "run", "", ErrMissingRequiredArg
		}
		return "run", parts[1], nil
	case ":write", ":w":
		if len(parts) < 2 {
			return "write", "", ErrMissingRequiredArg
		}
		return "write", parts[1], nil
	case ":back", ":b":
		if len(parts) < 2 {
			return "back", "1", nil
		}
		return "back", parts[1], nil
	default:
		return parts[0], "", ErrInvalidCommand
	}
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

	if f, err := os.Open("/tmp/.cyclops_history"); err == nil {
		line.ReadHistory(f)
		f.Close()
	}

mainloop:
	for {
		input, err := line.Prompt(prompt + "> ")
		if err == io.EOF {
			fmt.Println() //Returns user to prompt on a new line
			break mainloop
		}
		command, args, err := parseCommand(input)
		if err != nil {
			fmt.Println(err, command)
			continue
		}
		switch command {
		case "help":
			help()
			continue
		case "quit":
			fmt.Println("Exiting...")
			break mainloop
		case "commit":
			if id, err := ws.CommitLast(); err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("Committed:", id)
			}
		case "eval":
			if res, err := ws.Eval(args, true); err != nil {
				fmt.Println(err)
			} else {
				printResults(res)
			}
		case "from":
			if err := ws.SetImage(args); err != nil {
				fmt.Println("error setting image:", err)
			} else {
				fmt.Println("Image: ", args)
			}
		case "back":
			num, err := strconv.Atoi(args)
			if err != nil {
				fmt.Println("Error: invalid number specified")
				break
			}
			if err := ws.back(num); err != nil {
				fmt.Println("Error:", err)
			} else {
				fmt.Printf("went back %d steps\n", num)
			}
		case "history":
			ws.PrintHistory()
		case "print":
			if out, err := ws.Sprint(); err == nil {
				for _, line := range out {
					fmt.Println(line)
				}
			}
		case "run":
			if res, err := ws.Run(args); err != nil {
				fmt.Println(err)
			} else {
				printResults(res)
			}
		case "write":
			if args == "" {
				fmt.Println("Missing file path: `:write [path/to/file]`")
				break
			}
			if err := ws.Write(args); err != nil {
				fmt.Println("Error writing file:", err)
			} else {
				fmt.Println("File written:", args)
			}
		default:
			continue
		}

		line.AppendHistory(input)
		if f, err := os.Create("/tmp/.cyclops_history"); err != nil {
			fmt.Println("error writing history:", err)
		} else {
			line.WriteHistory(f)
			f.Close()
		}
	}
}
