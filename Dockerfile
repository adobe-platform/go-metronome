FROM       alpine:3.4
MAINTAINER BehanceRE <qa-behance@adobe.com>


# add dependencies
RUN apk add --no-cache \
      bash \
      curl \
      openssh-client 

# install runtime scripts
ADD metronome-cli-linux-amd64 /usr/local/bin/metronome-cli
 
CMD /usr/local/bin/metronome-cli 

