FROM alpine:3.3
MAINTAINER admin@aiicy.com

# Install system utils & runtime dependencies
ADD https://github.com/tianon/gosu/releases/download/1.10/gosu-amd64 /usr/sbin/gosu
RUN chmod +x /usr/sbin/gosu \
 && apk --no-cache --no-progress add ca-certificates bash git s6 curl socat

COPY . /app/aiicyds/
WORKDIR /app/aiicyds/
RUN ./docker/build.sh

# Configure LibC Name Service
COPY docker/nsswitch.conf /etc/nsswitch.conf

# Configure Docker Container
VOLUME ["/data/aiicyds"]
EXPOSE 5555
ENTRYPOINT ["docker/start.sh"]
CMD ["/bin/s6-svscan", "/app/aiicyds/docker/s6/"]