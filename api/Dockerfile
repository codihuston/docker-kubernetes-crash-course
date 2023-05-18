# Source: https://hub.docker.com/_/golang
FROM golang:1.20.4-bullseye

# Where our application will live in the completed container
WORKDIR /src

# Copy dependencies such as package manager manifests to our WORKDIR
# Note: the context of copy directives is relative to the WORKDIR.
# i.e.) These files are copied into /src/go.mod, etc.
COPY go.mod go.sum ./

# Install dependencies
RUN go mod download

# I want our container to remain online while we are developing
CMD ["sleep", "infinity"]
