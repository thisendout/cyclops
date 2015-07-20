package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

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
:h, :help                     show help
:f, :from      [image]        set base image
:e, :eval      [command ...]  execute shell command (ephemeral)
:r, :run       [command ...]  execute shell command (auto commits image)
:c, :commit                   commit changes from last command
:b, :back      [num]          go back in the history (default: 1)
:hs, :history                 show the current history
:w, :write     [path/to/file] write state to file
:q, :quit                     quit cyclops - <ctrl-d>
`
	fmt.Println(usage)
}

func printResults(res EvalResult) {
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

func printHistory(history []EvalResult, currentImage string) {
	n := 1

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 0, '\t', 0)

	// print header
	fmt.Fprintln(w, "Item\tCommand\tExit\tCreated Image")

	for _, entry := range history {
		var row string
		if entry.Deleted {
			row = "\t"
		} else {
			if currentImage == entry.NewImage {
				row += fmt.Sprintf(">%d\t", n)
			} else {
				row += fmt.Sprintf("%2d\t", n)
			}
			n += 1
		}
		row += fmt.Sprintf("%s\t", entry.Command)
		row += fmt.Sprintf("%d\t", entry.Code)
		if len(entry.NewImage) == 64 {
			row += fmt.Sprintf("%s\t", entry.NewImage[:12])
		} else {
			row += fmt.Sprintf("%s\t", entry.NewImage)
		}
		fmt.Fprintln(w, row)
	}
	w.Flush()
}

func pruneChanges(changes []docker.Change) []docker.Change {
	var p string
	c := []docker.Change{}
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

func preExit(ws *Workspace) {
	fmt.Println("Cleaning up...")
	lines := ws.Reset()
	for _, line := range lines {
		if line.Err != nil {
			fmt.Println(line.Err)
		} else {
			fmt.Printf("Deleted: %s\n", line.Id)
		}
	}
	fmt.Println("Done")
}

func main() {
	dc, err := NewDockerClient(os.Getenv("DOCKER_HOST"), os.Getenv("DOCKER_TLS_VERIFY"), os.Getenv("DOCKER_CERT_PATH"))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if err := dc.Ping(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("Connected to docker daemon...")

	ws := NewWorkspace(dc, "bash", defaultImage)

	line := liner.NewLiner()
	defer line.Close()

	prompt := defaultPrompt

	if f, err := os.Open("/tmp/.cyclops_history"); err == nil {
		line.ReadHistory(f)
		f.Close()
	}

	defer preExit(ws)

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
			break mainloop
		case "commit":
			if id, err := ws.CommitLast(); err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("Committed:", id)
			}
		case "eval":
			if res, err := ws.Eval(args); err != nil {
				fmt.Println(err)
			} else {
				printResults(res)
			}
		case "from":
			if ws.CurrentImage != ws.Image {
				confirm, err := line.Prompt("Changes will be lost. Continue? <y>: ")
				if err != nil || confirm != "y" {
					fmt.Println("Aborted")
					continue
				}
				fmt.Println("Wiping history to set new base image")
			}
			ws.Reset()
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
				fmt.Printf("Back %d to %s\n", num, ws.CurrentImage)
			}
		case "history":
			printHistory(ws.history, ws.CurrentImage)
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
