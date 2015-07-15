# sysrepl

A docker-backed [read-eval-print loop](https://en.wikipedia.org/wiki/Read%E2%80%93eval%E2%80%93print_loop) for systems scripting and configuration.

sysrepl executes commands inside of a docker container and provides immediate feedback.

## Getting Started

We're not packaging binaries just yet.  You'll need a Go environment setup and then:

```
$ go get github.com/thisendout/sysrepl
```

Make sure you have a DOCKER_HOST env variable set. [docker-machine](https://github.com/docker/machine) can help with this part.

```
$ docker-machine env dev
export DOCKER_TLS_VERIFY=yes
export DOCKER_CERT_PATH=/home/sysrepl/.docker/machine/machines/dev
export DOCKER_HOST=tcp://192.168.99.100:2376
$ docker pull ubuntu:trusty
```

Launch the repl and try some bash commands.

```
$ sysrepl
sysrepl> :help
sysrepl> :run apt-get update -y
sysrepl> :run apt-get install -y tmux
sysrepl> :print
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

* ```:print``` - Prints the source/commands run in the session formatted for the session type.

* ```:write filename``` - Writes the source/commands to a file given the session type.

## License

sysrepl is released under the MIT License (c) 2015 This End Out, LLC. See `LICENSE` for the full license text.
