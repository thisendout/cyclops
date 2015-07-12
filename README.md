```
# Given a machine named 'dev'
$(docker-machine env dev)
docker build -t ansible .
go install
sysrepl
:image ansible
:ansible site.yml
```
