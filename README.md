# cyclops
[![Circle CI](https://circleci.com/gh/thisendout/cyclops.svg?style=svg)](https://circleci.com/gh/thisendout/cyclops)

A docker-backed [read-eval-print loop](https://en.wikipedia.org/wiki/Read%E2%80%93eval%E2%80%93print_loop) for systems scripting and configuration.  Faster than a VM, safer than production, more feedback than a local terminal.

cyclops executes commands inside of a docker container and provides immediate feedback.  Commit changes to incrementally build a final image or Dockerfile.  Quickly roll back if you don't like what you see.  When you're done, write out your history to a Dockerfile and/or save the resulting docker container for later.

cyclops is designed to help with:
 * exploring shell commands in a safe, quick manner
 * building Dockerfiles
 * testing the execution of scripts and automation

## Getting Started

We're not packaging binaries just yet.  You'll need a Go environment setup and then:

```
$ go get github.com/thisendout/cyclops
```

Make sure you have a DOCKER_HOST env variable set. [docker-machine](https://github.com/docker/machine) can help with this part.

```
$ docker-machine env dev
export DOCKER_TLS_VERIFY=yes
export DOCKER_CERT_PATH=/home/cyclops/.docker/machine/machines/dev
export DOCKER_HOST=tcp://192.168.99.100:2376
$ docker pull ubuntu:trusty
```

Launch the repl and try some bash commands.

```
$ cyclops
cyclops> :help
cyclops> :run apt-get update -y
cyclops> :run apt-get install -y tmux
cyclops> :print
```

### Workflows

cyclops aims to be flexible in how you explore and commit changes to your environment.

Entering a command into the cyclops prompt will execute the command without committing the change.  Use it for exploration and testing.

```
cyclops> dpkg -l | grep tmux     # no tmux installed
cyclops> apt-cache search tmux
...
cyclops> apt-get install -y tmux
...
cyclops> dpkg -l | grep tmux     # no tmux installed because apt-get install was ephemeral
```

`:commit` commits the change from the previous execution, regardless of exit code.
```
cyclops> dpkg -l | grep tmux     # no tmux installed
cyclops> apt-cache search tmux
...
cyclops> apt-get install -y tmux
...
cyclops> :commit                 # commit the changes from the previous step
cyclops> dpkg -l | grep tmux     # tmux is now installed
```

`:run` auto-commits the change if the command returns with exit `0`.
```
cyclops> dpkg -l | grep tmux     # no tmux installed
cyclops> :run apt-get install -y tmux
...
cyclops> dpkg -l | grep tmux     # tmux is now installed
```
If a `:run` returns non-zero but you want to commit it anyway, follow it up with `:commit`.

When you're done, use `:print` and `:write` to get a Dockerfile representing your commit changes.  cyclops also prints the container and image ids at every step if you want to use the resulting artifacts directly.

## Commands

* ```:from name``` - Accepts a single argument of a Docker image name to use for execution of commands.
  * Consumers:
    * Dockerfile (```FROM```)

* ```:run command``` - Runs a shell command against an image and displays the STDOUT and filesystem diff.
  * Consumers:
    * Dockerfile (```RUN```)

* ```:commit``` - Commits the container created from the previous command and uses it as the base image for the next command.

* ```:print``` - Prints the source/commands run in the session formatted for the session type.

* ```:write filename``` - Writes the source/commands to a file given the session type.

* All other entered commands are executed against the current image and results are displayed, but the changes are not committed.  You can `:commit` the change for the previous run, if desired.  Use bare commands to experiment or explore the current environment.

## Output

For each :run executed, cyclops reports:
 * Exit Code
 * Execution duration
 * Docker image used as base
 * Committed docker image ID with changes (if the changes were committed)
 * List of filesystem changes


## License

cyclops is released under the MIT License (c) 2015 This End Out, LLC. See `LICENSE` for the full license text.
