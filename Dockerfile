FROM       golang:1.7.3-alpine

# install runtime scripts
ADD . $GOPATH/src/github.com/adobe-platform/go-metronome
WORKDIR $GOPATH/src/github.com/adobe-platform/go-metronome


RUN apk add --no-cache \
      bash \
      build-base \
      curl \
      make \
      git \
    && make install-deps install 

 
CMD /usr/local/bin/skopos 

