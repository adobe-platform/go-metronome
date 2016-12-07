FROM       golang:1.7.3-alpine

# install runtime scripts
ADD . $GOPATH/src/github.com/adobe-platform/go-metronome
WORKDIR $GOPATH/src/github.com/adobe-platform/go-metronome


RUN apk add --virtual .pbbuild --no-cache \
      bash \
      build-base \
      curl \
      make \
      git \
    && make install-deps compile \
    && cp  metronome-cli-linux-amd64 /usr/local/bin
    && chmod +x /usr/local/bin/metronome-cli-linux-amd64  
 
CMD /usr/local/bin/metronome-cli-linux-amd64 

