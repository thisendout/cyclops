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
