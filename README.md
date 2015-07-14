# sysrepl

A [read-eval-print loop](https://en.wikipedia.org/wiki/Read%E2%80%93eval%E2%80%93print_loop) for system programming and tasks.

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

## Example
```
# Given a machine named 'dev'
$(docker-machine env dev)
docker build -t ansible .
docker pull ubuntu:trusty
go install
sysrepl
:image ansible
:ansible site.yml
:image ubuntu:trusty
:bash touch /tmp/foo
```
