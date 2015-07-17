# cyclops
[![Circle CI](https://circleci.com/gh/thisendout/cyclops.svg?style=svg)](https://circleci.com/gh/thisendout/cyclops)

A docker-backed [read-eval-print loop](https://en.wikipedia.org/wiki/Read%E2%80%93eval%E2%80%93print_loop) for systems scripting and configuration.  Faster than a VM, safer than production, more feedback than a local terminal.

cyclops executes commands inside of a docker container and provides immediate feedback.  Successful runs create a new base image used for subsequent runs automatically.  Quickly roll back if you don't like what you see.  When you're done, write out your history to a Dockerfile or shell script and/or save the resulting docker container for later.

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

## Commands

* ```:type arg``` - Sets the session to display/export it's state to a given format.
  * Arguments:
    * shell - export for the bash shell
    * dockerfile - export for a Dockerfile

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
 * Execution Duration
 * Docker image used as base
 * Committed docker image ID with changes (if exit was 0)
 * List of filesystem changes


## License

cyclops is released under the MIT License (c) 2015 This End Out, LLC. See `LICENSE` for the full license text.
