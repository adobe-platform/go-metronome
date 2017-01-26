DEV_VER=0.1


default: compile

install-deps:
	go get golang.org/x/tools/cmd/cover
	go get -u github.com/golang/lint/golint
	go get -u github.com/kardianos/govendor
	govendor sync

docker_test:
	@go test -v $$(go list ./... | grep -v /vendor/)

docker_vet:
	@go tool vet -all metronome metronome-cli/cli_support metronome-cli/ 

docker_lint:
	@for codeDir in metronome metronome-cli/cli_support metronome-cli/; do         LINT="$$(golint $$codeDir)" &&         if [ ! -z "$$LINT" ]; then echo "$$LINT" && FAILED="true"; fi; done && if [ "$$FAILED" = "true" ]; then exit 1; fi

# Make compilation depend on the docker dev container
# Run the build in the dev container leaving the artifact on completion
# Use run-dev to get an interactive session


docker_compile: docker_lint docker_vet
	make build-linux-amd64 build-darwin-amd64 

build-container: compile test
	@echo "Building go-metronome container ..."
	if [ "x$$sha" = "x" ] ; then sha=`git rev-parse HEAD`; fi ;\
	if [ ! -e /.dockerinit ]; then \
		docker build --tag adobe-platform/go-metronome:$$sha .; \
	else \
		echo "You're in a docker container. Leave to run docker" ;\
	fi

upload-current: build-container
	if [ "x$$sha" = "x" ] ; then sha=`git rev-parse HEAD`; fi ;\
	docker push adobe-platform/go-metronome:$$sha ; \
	docker tag adobe-platform/go-metronome:$$sha adobe-platform/go-metronome:latest ; \
	docker push adobe-platform/go-metronome:latest

build: compile


# run a shell in the docker dev environment, mounting this directory and establishing bash_history in the container instance
run-dev: dev-container
#       save bash history in-between runs...
	@if [ ! -f ~/.bash_history-go-metronome ]; then touch ~/.bash_history-go-metronome; fi
#       mount the current directory into the dev build
	docker run -i --rm --net host -e HISTSIZE=100000 -v ~/.bash_history-go-metronome:/root/.bash_history -v `pwd`:/go/src/github.com/adobe-platform/go-metronome -w /go/src/github.com/adobe-platform/go-metronome -t adobe-platform/go-metronome:dev bash


# build the docker dev container if it doesn't exists
dev-container:    ## makes container flotilla:1.7.3-dev and installs go deps
	@if [ ! -e /.dockerinit ]; then \
	  (docker images | grep 'adobe-platform/go-metronome' | grep -q dev) || \
	  docker build -f Dockerfile-dev -t adobe-platform/go-metronome:dev . ; \
	fi


# cross compilation works fine with 1.7.3.  using docker to ensure that

build-darwin-amd64: go-metronome-darwin-amd64

go-metronome-darwin-amd64: 
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.version=`git rev-parse HEAD`" -o metronome-cli-darwin-amd64 ./metronome-cli

build-linux-amd64: go-metronome-linux-amd64

go-metronome-linux-amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=`git rev-parse HEAD`" -o metronome-cli-linux-amd64 ./metronome-cli


resources compile lint test: dev-container
#   either ssh key or agent is needed to pull adobe-platform sources from git
#   this supplies to methods
#
	@if [ ! -e /.dockerinit ]; then \
	test -f ~/.bash_history-metronome || touch ~/.bash_history-metronome ;\
	SSH1="" ; SSH2="" ;\
	if [ ! -z "$$SSH_AUTH_SOCK" ] ; then SSH1="-e SSH_AUTH_SOCK=/root/.foo -v $$SSH_AUTH_SOCK:/root/.foo" ; fi ; \
	if [ -e ~/.ssh/id_rsa ]; then SSH2="-v ~/.ssh/id_rsa:/root/.ssh/id_rsa" ; fi ; \
	if [ "x$$sha" = "x" ] ; then sha=`git rev-parse HEAD`; fi ;\
	AWS=$$(env | grep AWS | xargs -n 1 -IXX echo -n ' -e XX') ;\
	docker run -i --rm $$SSH1 $$SSH2 $$AWS\
	-e sha=$$sha \
	-v ~/.bash_history-metronome:/root/.bash_history \
	-v $$(pwd):/go/src/github.com/adobe-platform/go-metronome \
	-w /go/src/github.com/adobe-platform/go-metronome \
	-t adobe-platform/go-metronome:dev \
		make docker_$@ ;\
	else \
		make docker_$@ ;\
	fi

run:
	@if [ ! -z "$(http_proxy)"]; then export PROXY="-e http_proxy=$http_proxy" ; fi ; echo docker run -i --rm $$PROXY --net host -t adobe-platform/go-metronome:`git rev-parse HEAD` /usr/local/bin/go-metronome-cli-linux-amd64
