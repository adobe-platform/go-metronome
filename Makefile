DEV_VER=0.1

default: compile

install-deps:
	go get -u github.com/kardianos/govendor
	govendor sync

test:
	make install-deps 
	go vet ./...
	golint ./...
	ginkgo -r -trace -failFast -v --cover --randomizeAllSpecs --randomizeSuites -p
	echo "" && for i in $$(ls **/*.coverprofile); do echo "$${i}" && go tool cover -func=$${i} && echo ""; done
	echo "" && for i in $$(ls **/**/*.coverprofile); do echo "$${i}" && go tool cover -func=$${i} && echo ""; done


# Make compilation depend on the docker dev container
# Run the build in the dev container leaving the artifact on completion
# Use run-dev to get an interactive session
docker-compile: dev
	@test -f ~/.bash_history-metronome || touch ~/.bash_history-metronome
	docker run -i --rm --net host -v ~/.bash_history-metronome:/root/.bash_history -v `pwd`:/go/src/github.com/adobe-platform/go-metronome -w /go/src/github.com/adobe-platform/metronome -e version=0.0.1  -e CGO_ENABLED=0 -e GOOS=linux -t adobe-platform/go-metronome:dev make compile


build-container: compile
	@echo "Building go-metronome container ..."
	@if [ ! -e /.dockerinit ]; then \
		docker build --tag adobe-platform/go-metronome:`git rev-parse HEAD` .; \
	else \
		echo "You're in a docker container. Leave to run docker" ;\
	fi

upload-current:
	make build-container
	docker push adobe-platform/metronome:`git rev-parse HEAD`
	docker tag adobe-platform/go-metronome:`git rev-parse HEAD` adobe-platform/go-metronome:latest
	docker push adobe-platform/go-metronome:latest

build: compile

# build the docker dev container if it doesn't exists
dev:
	@if [ ! -e /.dockerinit ]; then \
	  (docker images | grep 'adobe-platform/go-metronome' | grep -q dev) || \
	  docker build -f Dockerfile-dev -t adobe-platform/go-metronome:dev . ; \
	fi

# run a shell in the docker dev environment, mounting this directory and establishing bash_history in the container instance
run-dev: dev_container
#       save bash history in-between runs...
	@if [ ! -f ~/.bash_history-go-metronome ]; then touch ~/.bash_history-go-metronome; fi
#       mount the current directory into the dev build
	docker run -i --rm --net host -e HISTSIZE=100000 -v ~/.bash_history-go-metronome:/root/.bash_history -v `pwd`:/go/src/github.com/adobe-platform/go-metronome -w /go/src/github.com/adobe-platform/go-metronome -t adobe-platform/go-metronome:1.7.3-dev bash


# build the docker dev container if it doesn't exists
dev_container:    ## makes container flotilla:1.7.3-dev and installs go deps
dev_container:
	@grep -q docker  /proc/1/cgroup ; \
        if [ $$? -ne 0 ]; then \
	  (docker images | grep 'go-metronome' | grep -q 1.7.3-dev) || \
	  docker build -f Dockerfile-dev -t adobe-platform/go-metronome:1.7.3-dev . ; \
	fi


# cross compilation works fine with 1.7.3.  using docker to ensure that

build-darwin-amd64: go-metronome-darwin-amd64

go-metronome-darwin-amd64: 
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.version=`git rev-parse HEAD`" -o go-metronome-cli-darwin-amd64 ./metronome-cli

build-linux-amd64: go-metronome-linux-amd64

go-metronome-linux-amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=`git rev-parse HEAD`" -o go-metronome-cli-linux-amd64 ./metronome-cli


compile:           # cross compiles go-metronome producing darwin and linux ready binaries
compile: dev_container
	@grep -q docker  /proc/1/cgroup ; \
        if [ $$? -ne 0 ]; then \
		docker run -i --rm \
		-v $$(pwd):/go/src/github.com/adobe-platform/go-metronome \
		-w /go/src/github.com/adobe-platform/go-metronome \
                -t adobe-platform/go-metronome:1.7.3-dev \
		make build-linux-amd64 build-darwin-amd64 ;\
	else \
		make build-linux-amd64 build-darwin-amd64 ; \
	fi
run:
	@if [ ! -z "$(http_proxy)"]; then export PROXY="-e http_proxy=$http_proxy" ; fi ; echo docker run -i --rm $$PROXY --net host -t adobe-platform/go-metronome:`git rev-parse HEAD` /usr/local/bin/go-metronome-cli-linux-amd64
