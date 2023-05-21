# Source: https://hub.docker.com/_/golang
FROM golang:1.20.4-bullseye

# Where our application will live in the completed container
WORKDIR /api

# Copy all files, including dependency manifests
COPY . ./

# Build our application
RUN CGO_ENABLED=0 go build -a -installsuffix cgo -o server .

# Install dependencies
RUN go mod download
RUN apt-get update && \
    apt-get install -y \
    apt-transport-https \
    ca-certificates \
    curl \
    gnupg-agent

# Install golang migrate tool (/usr/bin/migrate)
RUN curl -sSL https://packagecloud.io/golang-migrate/migrate/gpgkey | apt-key add -
RUN echo "deb https://packagecloud.io/golang-migrate/migrate/debian/ bullseye main" > /etc/apt/sources.list.d/migrate.list
RUN apt-get update && \
    apt-get install -y migrate

# Note: we use a separate build stage to keep our docker image as slim as
# possible. Once our application is built in the previous stage, we take the
# binaries we care about, and put them in this final image below, since we
# no longer need the dependencies that were required to build our app in the
# first place.
FROM alpine:latest  
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=0 /api/server .
COPY --from=0 /api/db/migrations ./db/migrations
COPY --from=0 /usr/bin/migrate /usr/bin/migrate
COPY --from=0 /api/entrypoint.sh /run/entrypoint.sh

# Run our server on boot
CMD ["/run/entrypoint.sh"]
