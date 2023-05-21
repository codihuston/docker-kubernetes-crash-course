# Purpose

This repository is a crash course on how to develop container-native
applications.

## Disclaimers

While this is intended to educate developers on development workflows for
container-native applications, much of the implementation is opinionated, and
not a golden standard of how to do things. However, wherever possible, I attempt
to provide context, existing standards or patterns, and other such references
to provide understanding for why a decision is being made.

For example, the quality of the code (software design patterns, glue code, etc.)
comes secondary to the experience that is intended to be gained here.

## Prerequisites

- [Docker Desktop](https://www.docker.com/products/docker-desktop/) or [Docker Engine for Linux](https://docs.docker.com/engine/install/ubuntu/)

  > Note: if you're on Docker v1 in Linux, you need to install
  > [Docker Compose](https://docs.docker.com/compose/install/linux/) separately.

- [KinD (Kubernetes-in-Docker)](https://kind.sigs.k8s.io/docs/user/quick-start/#installation)

  If you are using KinD, it is recommended to use it [with an image registry]().
  [I have a repository here](https://github.com/codihuston/kind-with-registry)
  that I use for deploying different versions of Kubernetes with an image
  registry. Currently, these scripts are designed to run one cluster and
  registry at a time. If you need to run multiple clusters and multiple
  registries, you will need to update the port mapping of the image registry
  for each instance. We will only use one instance for this lab.

- [Kubectl](https://kubernetes.io/docs/tasks/tools/)

  > Note: it is recommended that your `kubectl` client matches your Kubernetes
  > server major version, but this is not a strict requirement.

- [Delve](https://github.com/go-delve/delve)

  While we will be developing a go application within a container that
  contains all of the appropriate tooling, you will want to consider
  installing [Go](https://go.dev/dl/).

## Table of Contents

[Table of Contents]
- [Purpose](#purpose)
  - [Disclaimers](#disclaimers)
  - [Prerequisites](#prerequisites)
  - [Table of Contents](#table-of-contents)
  - [How to Use](#how-to-use)
  - [Topics](#topics)
  - [The Demo Application](#the-demo-application)
- [Docker](#docker)
  - [How to Run a Docker Container](#how-to-run-a-docker-container)
  - [How to Build Your Own Docker Container](#how-to-build-your-own-docker-container)
  - [Volume Mounts](#volume-mounts)
  - [Environment Variables and Executing Commands Within the Container](#environment-variables-and-executing-commands-within-the-container)
  - [Docker Networking (and DNS resolution)](#docker-networking-and-dns-resolution)
- [Docker Compose](#docker-compose)
  - [Prerequisites](#prerequisites-1)
  - [Initializing the API Application](#initializing-the-api-application)
  - [Dockerizing the API](#dockerizing-the-api)
  - [Adding the Database Layer](#adding-the-database-layer)
  - [Fix Broken Imports in VSCode](#fix-broken-imports-in-vscode)
  - [Adding CRUD Features to API](#adding-crud-features-to-api)
    - [Create a Blog Record](#create-a-blog-record)
    - [Get All Blogs](#get-all-blogs)
    - [Get a Single Blog](#get-a-single-blog)
    - [Update a Blog](#update-a-blog)
    - [Delete a Blog](#delete-a-blog)
  - [Setup the Go Debugger](#setup-the-go-debugger)
    - [Running Tests](#running-tests)
  - [Testing I](#testing-i)
    - [A Handful of Unit Tests](#a-handful-of-unit-tests)
    - [Unit Testing, Integration Testing, and Stubs I](#unit-testing-integration-testing-and-stubs-i)
    - [Unit Testing, Integration Testing, and Stubs II](#unit-testing-integration-testing-and-stubs-ii)
- [A Major Refactor - Business and Software Architecture](#a-major-refactor---business-and-software-architecture)
  - [Enterprise Architecture](#enterprise-architecture)
  - [Onion Architecture](#onion-architecture)
  - [SOLID Principals](#solid-principals)
    - [Unit Tests and Mocks](#unit-tests-and-mocks)
    - [A Simple Integration Test](#a-simple-integration-test)
    - [A Simple E2E Test](#a-simple-e2e-test)

## How to Use

This repository is provided as-is. It is recommened that you follow the
instructions below and execute them yourself in an empty directory, and use
this repository as a completed example.

> Pro tip: step through and review each commit in the git history.

## Topics

1. [Docker](#docker)
2. [Docker Compose](#docker-compose)
   1. Composing a microservice application
   2. Writing tests in Go
   3. Using the golang debugger
3. Kubernetes

   > Note: We'll cover these topics from the perspective of a developer who
   > simply wants to deploy their application in Kubernetes. Later, we'll
   > add more context that covers some more advanced topics.

   1. Primitives
      1. Pod
      2. Deployment
         1. Downward API
      3. ConfigMap
      4. Secrets
      5. Services
         1. Port Forwarding with Kubectl
   2. Architecture
      1. KubeAPI
      2. 
   3. Identity
      1. Service
      2. ServiceAccount
   4. Advanced Topics
      1. Volume Mounts - Sharing volumes between containers
      2. [Ingress](https://kind.sigs.k8s.io/docs/user/ingress/#ingress-nginx)

         Cover things like TLS termination, Sticky Sessions via annotations.

4. Kubernetes Operators
   
   This will be covered in a separate effort, but you can [start here](https://sdk.operatorframework.io/docs/building-operators/golang/tutorial/).

## The Demo Application

We will build a simple blog application. It will consist of the following
services and components:

1. REST API to CRUD a blog post
2. PostgreSQL back-end
3. Redis as a caching layer

The application will offer a **very insecure** method of authentication and
authorization for the purpose of demonstrating Ingress.

Again, the focus here is to demonstrate ways to be efficient in the developer
workflow for writing features.

# Docker

In this section, we'll experiment briefly with the most common features of
Docker. We will complete this section using the [nginx](https://hub.docker.com/_/nginx)
web server.

Containerization is a feature of the Linux Kernel. It enables the container
host to maintain an abstraction layer between the host and other containers, as
if the containers were run independently, agnostic of its host and other
containers. Containers can be networked together, so they can communicate to
each other.

> Note: these "wins" I'm listing aren't really in any particular order, they
> are just coming to mind as I type this...

The biggest win of containers is it does not require an entire hypervisor
like Virtual Machines do. This means that the virtualization is done at the
Host OS level, as opposed to the Hardware level (Type 1 Hypervisor) or the
Hypervisor on the Host OS level (Type 2). This is overall vastly more efficient
of compute resources.

The second biggest win is Docker Containers enable us to
"package software together" into an [OCI-compliant container](https://opencontainers.org/).
If you've ever configured a web server with PHP, PERL, or anything like that,
you know that it's a pain in the ass, and kind of a mystery. The advent
of [ansible](https://www.ansible.com/) and [chef](https://docs.chef.io/)
alleviated some of that overhead, but still requried overhead of its own.

However, a docker container image configured with PHP can be done once, and
distributed to anyone for use immediately with a container engine like Docker.
Apply this logic to any complex software, and boom, you can string together any
number of services far easier than ever before. For example, no more
[XAMPP](https://www.apachefriends.org/) or installing databases or web servers
directly onto your computer. You can deploy and throw away containers at will!

The third biggest win is precisely that, the disposability of running
containers. It makes it easy for you to deploy specific versions of a product,
like say `mysql` at will. All you have to do is change the `tag` of the docker
image and you can test your own product against a different version of `mysql`
with almost zero effort, allowing you discover gaps in your product sooner!
No more installing multiple versions of a database on your host, or on a
shared developer server!

The fourth biggest win is `developer experience`
(really, the biggest win for probably every *developer*)! If it works on my
machine, it will 99% work on your machine, and vastly faster than something
like [vagrant](https://www.vagrantup.com/). Caveats here depend on any glue
code I might run from my container host to initialize the environment--such
glue code may not work on your host (GNU vs BSD tools; Shell/Bash vs Windows
PowerShell, etc.).

Other container engines like `Podman`, or container runtimes exist, like
`dockerd`, `CRI-O`, `runc`, `crun`, etc. Container engines do a lot of things,
including abstracting away the runtime at a lower level, but mostly allow
us to interact with the contianers at a higher level, like concerns around
lifecycle management. Any OCI-compliant image should be able to run on any
container engine and runtime. There's a caveat here when it comes to running
containers as Rootless or Rootful users, but that is a topic of another day.

So let's get into it...

## How to Run a Docker Container

In a single terminal, run the following:

```bash
$ docker run nginx:1.25.0
Unable to find image 'nginx:1.25.0' locally
1.25.0: Pulling from library/nginx
# --- snip ---
# This will hang and print nginx server logs...
```

In another terminal, run the following:

```bash
$ docker ps
CONTAINER ID   IMAGE                  COMMAND                  CREATED          STATUS          PORTS                       NAMES
3be90bf1a88b   nginx:1.25.0           "/docker-entrypoint.…"   48 seconds ago   Up 48 seconds   80/tcp                      stupefied_gagarin
```

You can log a container as such:

> Note: use the `-f` switch to `tail` the log file (watch in realtime).

```bash
# using container id
$ docker logs 3be90bf1a88b

# using name
$ docker logs stupefied_gagarin
```

You may want to know more information about a container at runtime, such as
what command it runs at start, or its environment variables, or its volume
mounts, or its network information, base image information, resource limits,
etc.

```
$ docker inspect 3be90bf1a88b
```

Stop and remove the container:

```bash
# Stops the container. You can start it again if you want
$ docker stop 3be90bf1a88b

# Removes the container completely from docker
$ docker rm 3be90bf1a88b
```

> Note: removing the container from docker will destroy docker volumes
> associated with it. It will not destroy volume mounts that are mounted on the
> host.

When running a container, you can pass in a `--name`. You can also prevent
docker from keeping a handle on the container logs by specifying `--detach`.

```bash
$ docker run --name my-nginx --detach nginx:1.25.0
# Docker prints the container id with the --detach option
6bdb749eb78fc5efee30b539845ffa4253962f8a854a239089c658bdda26aef7
```

Then you can reference that container by name using the docker CLI:

```bash
# View the logs again
$ docker logs my-nginx

# Stop and remove the container
$ docker stop my-nginx && docker rm my-nginx
```

Read more here about [docker run](https://docs.docker.com/engine/reference/commandline/run/).

## How to Build Your Own Docker Container

You can build your own container from scratch (which is quite involved),
or based off an existing image. We will do the latter. You can use a base
image for almost any existing Linux operating system, or a pre-built container
from a vendor, which usually contains some arbitrary operating system with
their own software installed and configured with some defaults. That is
what we have done here with nginx.

A Dockerfile always starts with a `FROM` directive, which denotes the base
image we want to use for our new image.

From the root of this directory, change into the `nginx` directory and build
a container from the relative [Dockerfile](./nginx/Dockerfile). All we
are doing is baking in a file to this image. This directory
`/usr/share/nginx/html` is the default directory that `nginx` uses to serve
html files according to the [nginx dockerhub page](https://hub.docker.com/_/nginx).

```bash
# The "-f Dockerfile" is not required, and assumed by default. The "." is
# required, as it tells docker to build from this directory.
$ docker build -f Dockerfile -t custom-nginx .
[+] Building 0.1s (7/7) FINISHED
 => [internal] load build definition from Dockerfile                                                                                                     0.0s
 => => transferring dockerfile: 37B                                                                                                                      0.0s
 => [internal] load .dockerignore                                                                                                                        0.0s
 => => transferring context: 2B                                                                                                                          0.0s
 => [internal] load metadata for docker.io/library/nginx:1.25.0                                                                                          0.0s
 => [internal] load build context                                                                                                                        0.0s
 => => transferring context: 60B                                                                                                                         0.0s
 => [1/2] FROM docker.io/library/nginx:1.25.0                                                                                                            0.0s
 => CACHED [2/2] COPY ./html /usr/share/nginx/                                                                                                           0.0s
 => exporting to image                                                                                                                                   0.0s
 => => exporting layers                                                                                                                                  0.0s
 => => writing image sha256:1fa8235eaeac68ac3a86ecec90e627a8a04d1a5d23844b699394c49027af2ff6                                                             0.0s
 => => naming to docker.io/library/custom-nginx                                                                                                          0.0s
```

Now let's run our new nginx container, and map a port to it:

```bash
$ docker run -p 3000:80 --name my-nginx custom-nginx
```

In another terminal, query your docker container through the host mapping.

```bash
$ curl localhost:3000
<!DOCTYPE html>
Hello World!
</html>
```

Be sure to clean up the container when you're done:

```bash
$ docker stop my-nginx && docker rm my-nginx
```

There exists more directives, such as `RUN`, `ENTRYPOINT`, `CMD`, which are
commonly used. You can use the `RUN` command to install packages using the
operating system's package manager (yum/dnf for CentOS/RHEL-based images, apk
for alpine, and apt/deb for ubuntu/debian, etc).

Read more about the [Dockerfile](https://docs.docker.com/engine/reference/builder/).

##  Volume Mounts

When we mentioned that this is particularly useful for developers, this is
the feature that makes that possible. In essence, we can "mount" a directory
from the container host onto the container itself such that the files from
the host are visible from the container. This link allows us to sort of
"plug in" our files ad-hoc, which abstracts away the underlying software
dependency. This means that we can "bring our own files" to a specific version
of [go](https://hub.docker.com/_/golang) or any other language, and swapping
out the versions is simpler than ever.

It also allows us to propagate file changes into the container immediately,
without the need to copy/update the files within the container.

From the `nginx` directory, run the following:

> Note: we are using the absolute path to our `nginx/html` directory!

```
$ docker run -p 3000:80 --name my-nginx --volume $(pwd)/html:/usr/share/nginx/html custom-nginx
```

Then edit and save [nginx/html/index.html](./nginx/html/index.html). Query
the server to see that the file contents were in fact updated to whatever
you changed it to:

```
$ curl localhost:3000
```

Be sure to clean up the container when you're done:

```bash
$ docker stop my-nginx && docker rm my-nginx
```

## Environment Variables and Executing Commands Within the Container

Sometimes we will want to expose configuration to our container at the
environment level. Other options are to expose configuration via the file
system. We're concerned about the former.

```
$ docker run --name my-nginx -e MY_VAR="123" custom-nginx
```

Let's get a remote shell into the container:

> Note: not all docker containers have `bash` installed, so you might need
> to use `sh` instead.

```
$ docker exec -it my-nginx bash
root@5c602e5688b5:/#
```

Once you see `root@<container-id>`, you know that you're inside of the
container. Run the following now:

```bash
$ hostname
$ cat /etc/os-release
PRETTY_NAME="Debian GNU/Linux 11 (bullseye)"
NAME="Debian GNU/Linux"
VERSION_ID="11"
VERSION="11 (bullseye)"
VERSION_CODENAME=bullseye
ID=debian
HOME_URL="https://www.debian.org/"
SUPPORT_URL="https://www.debian.org/support"
BUG_REPORT_URL="https://bugs.debian.org/"

# Display your environment variables
$ env
# --- snip ---
MY_VAR=123
# --- snip ---

# Exit the session
$ exit
```

We can write our application code to accept values from environment variables
at runtime. Obviously, we'd want to handle exceptions where we expect some
environment variable to exist when it does not.

Be sure to clean up the container when you're done:

```bash
$ docker stop my-nginx && docker rm my-nginx
```

## Docker Networking (and DNS resolution)

Docker enables you to network containers together. This means that containers
can commmunicate to each other if they're on the same network. Network
containers automatically receive a DNS record matching the hostname. Let's
demonstrate that by deploying our own nginx container and the docker hub
nginx container on the same network.

```bash
$ docker network create my-network
3e79828e358e88643ed6f61409c44f4291173cfb58efeae8f43c2c135f0ec116

$ docker network ls
NETWORK ID     NAME                           DRIVER    SCOPE
# --- snip ---
3e79828e358e   my-network                     bridge    local
```

Let's deploy two containers on the same network:

```bash
# Using the nginx image we created
$ docker run --name nginx01 --network my-network --detach custom-nginx
47c3edf63e3aa08559f6647dbb7c7082b029d91a1201005ae3e4ee82b06295f6

# Using the nginx image from docker hub
$ docker run --name nginx02 --network my-network --detach nginx:1.25.0
a47db5bab6ee831dba7ae2f4d08abfa8450406006c66ec33ddec697a4726d730

$ docker ps
CONTAINER ID   IMAGE                  COMMAND                  CREATED         STATUS         PORTS                       NAMES
47c3edf63e3a   nginx:1.25.0           "/docker-entrypoint.…"   6 seconds ago   Up 6 seconds   80/tcp                      nginx02
a47db5bab6ee   nginx:1.25.0           "/docker-entrypoint.…"   9 seconds ago   Up 9 seconds   80/tcp                      nginx01
```

Exec into `nginx01` and query `nginx02`. We should receive the default nginx
landing page in the response body:

```bash
$ docker exec -it nginx01 curl nginx02
<!DOCTYPE html>
<html>
<head>
<title>Welcome to nginx!</title>
<style>
html { color-scheme: light dark; }
body { width: 35em; margin: 0 auto;
font-family: Tahoma, Verdana, Arial, sans-serif; }
</style>
</head>
<body>
<h1>Welcome to nginx!</h1>
<p>If you see this page, the nginx web server is successfully installed and
working. Further configuration is required.</p>

<p>For online documentation and support please refer to
<a href="http://nginx.org/">nginx.org</a>.<br/>
Commercial support is available at
<a href="http://nginx.com/">nginx.com</a>.</p>

<p><em>Thank you for using nginx.</em></p>
</body>
</html>
```


Now exec into `nginx02` and query `nginx01`. We should receive the default
webpage we baked into our nginx image:

```bash
$ docker exec -it nginx02 curl nginx01
<!DOCTYPE html>
Hello World!
</html>
```

Cleanup:

```bash
$ docker stop nginx01 && docker rm nginx01
$ docker stop nginx02 && docker rm nginx02
$ docker network rm my-network
```

At this point, you've used Docker for about 75% of what most users use it for.
There remaining exercises are to maintain dependencies in the `Dockerfile`,
using [Build Stages](https://docs.docker.com/build/building/multi-stage/), and
configuring the `ENTRYPOINT` or `CMD`. We cover most of this to some degree
later in this lab.

# Docker Compose

## Prerequisites

A [Docker Compose](https://docs.docker.com/compose/compose-file/compose-file-v3/)
project consists of a `docker-compose.yaml` file that describes a set of
services and their configurations. This would consist of a docker image
(built from scratch, or a pre-baked one from online, like docker hub), build
arguments, a set of environment variables, volume mounts, networking properties
(like a static ip address or shared network) and more.

`Docker Compose` makes environments portable. If it works on docker on one
machine, it should work on docker on another machine. Your mileage may vary
here, as there are several factors that may impact the truthfulness of the
previous statement. Things that may cause this variance might be:

1. The version of which images are used (for example, if you're using 3rd party
   images with the `latest` tag, newer containers may behave differently than
   older ones)

2. The version of docker you are using
3. Whether you are using Docker Engine on Linux, or Docker Desktop for MacOS or
   Windows

   Permissions are automatically mapped to the container host user (your user)
   in Docker Desktop, but that is not the case in Linux. For example, if you
   mount and run a container on a directory in your Linux host, and create a
   file in that volume from with in that container, your user will not have
   permissions on that file. Here are solution around this issue: [Avoiding Permission Issues With Docker-Created Files](https://vsupalov.com/docker-shared-permissions/).
   I would personally recommend using either chmod, or passing your `$UID` into
   the container environment. `docker run -e UID="$UID" ...`

4. We wary of glue code

   Sometimes, dev environments will leverage bash scripts (or others) that
   will wrap around Docker Compose. This might be a requirement in order to
   setup integrations with a complex product. Be sure that your workstation
   has an appropriate version of `bash` (v4+ associative arrays) or whatever
   language is used for your glue code.

5. ... and many other things

## Initializing the API Application

First, we will create the [api/go.mod](./api/go.mod) file. This is where
our golang dependencies will be stored.

```bash
$ mkdir api
$ cd api
$ go mod init example.com/m/v2
```

The repository that you specify is not important for this project, as it can
be anything you want. Just know that this is will be key in how you reference
packages (locally) that you develop in this project.

Install our [gin](https://github.com/gin-gonic/gin) dependency, a web framework
for golang:

```bash
$ go get github.com/gin-gonic/gin
```

Next, create `api/main.go` and add the following content:

```go
package main

import (
  "net/http"

  "github.com/gin-gonic/gin"
)

func main() {
  r := gin.Default()
  r.GET("/ping", func(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
      "message": "pong",
    })
  })
  r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
```

Run the project locally:

```bash
go run main.go

# Output
[GIN-debug] [WARNING] Creating an Engine instance with the Logger and Recovery middleware already attached.

[GIN-debug] [WARNING] Running in "debug" mode. Switch to "release" mode in production.
 - using env:   export GIN_MODE=release
 - using code:  gin.SetMode(gin.ReleaseMode)

[GIN-debug] GET    /ping                     --> main.main.func1 (3 handlers)
[GIN-debug] [WARNING] You trusted all proxies, this is NOT safe. We recommend you to set a value.
Please check https://pkg.go.dev/github.com/gin-gonic/gin#readme-don-t-trust-all-proxies for details.
[GIN-debug] Environment variable PORT is undefined. Using port :8080 by default
[GIN-debug] Listening and serving HTTP on :8080
```

In a separate tab in your terminal, run the following command to query the
API server:

```bash
curl localhost:8080

# Output
404 page not found

# Log output in the gin application
[GIN] 2023/05/17 - 23:01:08 | 404 |         600ns |       127.0.0.1 | GET      "/"
```

You can kill the server with `ctrl+c` in the terminal that you ran the
`go run` command in.

If you've made it this far, then you've successfully initialized a golang
web application, hosted it locally on your workstation, and queried it to
see that it is working!

Key takeaways:

- Dependencies for the developer

  In order to run our golang application, we had to install golang and packages
  associated with our application. In the real world, applications can become
  very complex, with a large number of dependencies and tooling. It can be
  difficult to get these things installed sometimes, which can lead to the
  "it works on my machine" meme. This is one of the key areas that docker
  improves quality of life.

In the next section, we'll move our project to a docker compose project from
which we will develop.

## Dockerizing the API

Let's create `api/Dockerfile`. This will contain all of the tooling that our
developers need in order to develop on the api application, notably `golang`.
Add the following content:

```bash
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
```

Next, let's create `docker-compose.yaml`, which will drive our application
lifecycle using the above docker image:

```bash
version: "3.9"
services:
  api:
    build:
      # The directory of which a target dockerfile exists
      context: ./api
      #dockerfile: Dockerfile-alternate # Docker Compose finds Dockerfile by
                                        # default. If you had other Dockerfiles,
                                        # this is how you'd specify them.
    ports:
      - "8080:8080"
    volumes:
       - ./api:/src
```

Run our dockerized application from the root of this repository (not the `api`
directory).

> Note: older versions of docker compose use the command `docker-compose`.
> It is recommended to use the new syntax where compose is built into the
> `docker` command, as shown below.

> Note: if you make changes to your Dockerfile,
```bash
# This command is blocking, and you will see all container output in the console
$ docker compose up

# You can run in detached mode if you prefer to free up your terminal tab
# (or --detach)
$ docker compose up -d
# ... and can access the container logs for this project as such
$ docker compose logs
```

In a separate terminal, confirm that the container is running:

```
$ docker ps
CONTAINER ID   IMAGE                                COMMAND                  CREATED          STATUS          PORTS                       NAMES
251d69f20148   docker-kubernetes-crash-course_api   "sleep infinity"         11 seconds ago   Up 10 seconds   0.0.0.0:8080->80/tcp        docker-kubernetes-crash-course_api_1
```

Note that the `port configuration` is mapping port `8080` from your workstation
to port `8080` of the docker container.  Confirm that the application is
running by hitting it with curl:

```bash
$ curl localhost:8080/
curl: (52) Empty reply from server
```

Oh right, our `api` server is not running inside of the container yet. Let's fix
that:

```bash
# From the root of this repository (same directory as docker-compose.yaml)
$ docker compose exec api bash

# Output: you get an ssh shell into your container
root@3274a566c7fa:/src#

# Run the application from inside the container
root@3274a566c7fa:/src# go run main.go
```

Once the server is running in the container, from another tab in your terminal
on your workstation (outside of the docker container) see if we can query the
server:

```bash
$ curl localhost:8080/

# Output
404 page not found
```

You can stop the compose environment by hitting `ctrl+c`. This will stop
the services defined in the compose file. IIf you are running in detached mode,
run `docker-compose down` from the root of this project instead. This isn't
obvious at this point in time, but these will also preserve any volume mount
data unless you specify the `-v or --volumes` argument, which will destroy
any docker volumes that are *not* attached to your host (like our `api`
directory is). We have not specified any volumes like this yet.

At this point, you have dockerized our application. If you were to collaborate
with other developers on this project, you can rest assured that if they have
`docker compose`, they should be able to get this exact same environment
replicated in their environment.

Key takeaways:

- The combination of your code, Dockerfile, and docker compose enable you to
  reproducable an environment exactly the same across devices
- Port mapping between your host and your container just gives you a network
  path from your `localhost` to your container. Do not forget to run the
  application within your container that listens on that port!
  
  In this environment, that is a requirement. In a production docker image, we
  might automatically start the server (preferred), or offer a CLI in-container
  to configure and start our server.
  
  In our developer environment, we could have some kind of process watch
  our filesystem for changes, and kill and recompile/re-run our application.
  [Air](https://github.com/cosmtrek/air) seems to be a promising tool to do
  such a thing. This would save you from having to kill/and re-run the
  `go run` command after making changes.

## Adding the Database Layer

In this section, we'll add [postgres](https://hub.docker.com/_/postgres) and
initialize the table using for our blogs

Create an environment variable file in the root of the repo called `.env`.
We will tell docker compose to source this for our `api` service.

> Note: In the real world, if you are not using a secrets manager to protect
> and distribute secrets (recommended), then you might fallback on using a
> `.env` file for each of your environments (prod, etc.). In development it is
> common to provide a `.env-example`, and to expect your developers to clone
> that to `.env` and to provide their own values for sensitive fields, such as
> an API Token to an external service.

```bash
$ touch .env
```

Update `docker-compose.yaml` to include this change, and add the `db` service,
which will host our postgresql server:

```diff
version: "3.9"
services:
  api:
    build:
      # The directory of which a target dockerfile exists
      context: ./api
      #dockerfile: Dockerfile-alternate # Docker Compose finds Dockerfile by
                                        # default. If you had other Dockerfiles,
                                        # this is how you'd specify them.
    ports:
      - "8080:8080"
    volumes:
       - ./api:/src
+    env_file: .env
+  db:
+    image: postgres:15.3
+    restart: always
+    environment:
+      POSTGRES_USER: postgres
+      POSTGRES_PASSWORD: postgres
+      POSTGRES_DB: blogger

```

Add the postgres connection string as an environment variable to the
`.env` file we created:

> Note: this connection string is connecting via the postgresql protocol,
> using the default `username:password` configured by the `POSTGRES_USER` and
> `POSTGRES_PASSWORD` environment variables, which are used to configure the
> postgres image [as per the docs](https://hub.docker.com/_/postgres#:~:text=on%20container%20startup.-,POSTGRES_PASSWORD,-This%20environment%20variable).
> We also in set the initial database name to `blogger`. We are disabling SSL
> because connections use that by default--enabling this is a separate exercise
> that we will not cover in this lab. That would be a requirement for
> production.

```
POSTGRESQL_URL=postgres://postgres:postgres@db:5432/blogger?sslmode=disable
```

Start the project:

```bash
$ docker-compose up
```

Let's verify that we can connect to the database from the new `db`
container:

```bash
$ docker compose exec db bash
root@bb9a81cc65d6:/# psql -U postgres
psql (15.3 (Debian 15.3-1.pgdg110+1))
Type "help" for help.

postgres=#
```

If you see the shell prompt `postgres=#` that means that the database is at
least running. Enter `\q` to exit the psql cli. Exit the container.

> Note: accessing the database this way usually means we've authenticated
> via the postgresql unix socket, which usually only confirms with the OS that
> the logged-in user is in `pg_hba.conf` with a `local auth-method`. If so,
> a password is not required. This is typical default behavior for the `root`
> user. Read more about [pg_hba.conf](https://www.postgresql.org/docs/current/auth-pg-hba-conf.html).
> We'll authenticate to the database from outside of the container soon, which
> will require a password.

When developing an application that uses a database, it's ideal to use a
migration tool so that you can change the database as you iterate on the
product. We opt for [golang-migrate/migrate](https://github.com/golang-migrate/migrate/tree/master)
in this project. Let's add the `migrate` tool to our api image:

> Note: we will use the default database as opposed to creating our own
> database. If you wanted to forego migrations, or configure your database
> at container startup, refer to [Initialization Scripts](https://hub.docker.com/_/postgres#:~:text=and%20POSTGRES_DB.-,Initialization%20scripts,-If%20you%20would)
> on the [postgres dockerhub page](https://hub.docker.com/_/postgres).

```diff
# Source: https://hub.docker.com/_/golang
FROM golang:1.20.4-bullseye

# Where our application will live in the completed container
WORKDIR /src

# Copy dependencies such as package manager manifests to our WORKDIR
# Note: the context of copy directives is relative to the WORKDIR.
# i.e.) These files are copied into /src/go.mod, etc.
COPY go.mod go.sum ./

# Install dependencies
+RUN go mod download
+RUN apt-get update && \
+    apt-get install -y \
+    apt-transport-https \
+    ca-certificates \
+    curl \
+    gnupg-agent
+
+# Install golang migrate tool
+RUN curl -sSL https://packagecloud.io/golang-migrate/migrate/gpgkey | apt-key add -
+RUN echo "deb https://packagecloud.io/golang-migrate/migrate/debian/ bullseye main" > /+etc/apt/sources.list.d/migrate.list
+RUN apt-get update && \
+    apt-get install -y migrate

# I want our container to remain online while we are developing
CMD ["sleep", "infinity"]
```

Restart the compose project by bringing it down and up again. Verify
that the `migrate` tool works from the `api` container:

> Note: if your `api` docker image does not include this tool, you can force
> docker compose to rebuild the container images `docker compose up --build`.

```bash
$ docker compose exec api bash
$ migrate -version
4.15.2
```

From the `api` container, create our migration files:

```bash
$ mkdir migrations
$ migrate create -ext sql -dir db/migrations -seq create_users_table
$ migrate create -ext sql -dir db/migrations -seq create_blogs_table
```

Populate the migration file contents as such:

> Important: if you are using docker on linux, you may not have permissions
> on these files. Docker Desktop automatically resolves these issues, but
> docker engine on linux does not, thus these files (created from the context
> within the container), will be owned by root. You can fix this by
> running the following on your workstation: `sudo chown -R $USER api`.
> Otherwise, you can prepend the following variables to your docker compose 
> commands: `UID="$(id -u)" GID="$(id -g)" docker-compose ...`. There are
> also [other solutions](https://devcoops.com/docker-compose-uid-gid/).

```sql
-- 000001_create_users_table.up.sql
CREATE TABLE IF NOT EXISTS users(
   id serial PRIMARY KEY,
   username VARCHAR (50) UNIQUE NOT NULL,
   password VARCHAR (50) NOT NULL,
   email VARCHAR (300) UNIQUE NOT NULL,
   created_at TIMESTAMPTZ,
   updated_at TIMESTAMPTZ,
   deleted_at TIMESTAMPTZ
);

```

```sql
-- 000001_create_users_table.down.sql
DROP TABLE IF EXISTS users;
```

```sql
-- 000002_create_blogs_table.down.sql
CREATE TABLE IF NOT EXISTS blogs(
   id serial PRIMARY KEY,
   title VARCHAR (50) UNIQUE NOT NULL,
   body TEXT NOT NULL,
   created_at TIMESTAMPTZ,
   updated_at TIMESTAMPTZ,
   deleted_at TIMESTAMPTZ
);

```

From the `api` container, run the migrations that we created.

> Note: this is making use of the `POSTGRESQL_URL` variable we defined in our
> `.env` file.

```bash
$ migrate -database ${POSTGRESQL_URL} -path db/migrations up
# Output
1/u create_users_table (12.7821ms)
2/u create_blogs_table (25.011ms)
```

From the `db` container, verify the tables exist:

```bash
# sign into postgres
$ psql -U postgres

# connect to the bloggers database
postgres= \c blogger
postgres= \dt
               List of relations
 Schema |       Name        | Type  |  Owner
--------+-------------------+-------+----------
 public | blogs             | table | postgres
 public | schema_migrations | table | postgres
 public | users             | table | postgres
 
postgres= \d users
                                       Table "public.users"
  Column  |          Type          | Collation | Nullable |                Default
----------+------------------------+-----------+----------+----------------------------------------
 user_id  | integer                |           | not null | nextval('users_user_id_seq'::regclass)
 username | character varying(50)  |           | not null |
 password | character varying(50)  |           | not null |
 email    | character varying(300) |           | not null |
Indexes:
    "users_pkey" PRIMARY KEY, btree (user_id)
    "users_email_key" UNIQUE CONSTRAINT, btree (email)
    "users_username_key" UNIQUE CONSTRAINT, btree (username)

postgres= \d blogs
                                      Table "public.blogs"
 Column  |         Type          | Collation | Nullable |                Default
---------+-----------------------+-----------+----------+----------------------------------------
 blog_id | integer               |           | not null | nextval('blogs_blog_id_seq'::regclass)
 title   | character varying(50) |           | not null |
 body    | text                  |           | not null |
Indexes:
    "blogs_pkey" PRIMARY KEY, btree (blog_id)
    "blogs_title_key" UNIQUE CONSTRAINT, btree (title)
```

At this point, you have successfully configured a postgres database service
in our docker compose project, and connected to it with a client from another
container.

> Important: at this point, we are not concerned authenticating or authorizing
> end-users, and associating them with blog posts. That may come in a later
> exercise.

Key takeaways:

- Containers in a compose project can automatically communicate to each other
  via a DNS record matching their service name

  Our `api` service can connect to our postgres database using `db` as the
  DNS name. This means that anytime we have a client that needs to connect
  to a service that is hosted in another container in this project, we can
  simply specify the service name as the `host` or `hostname`.

- You learned how to configure a `postgres` docker container and how to
  configure the initial database, username, and password

- You learned about [pg_hba.conf](https://www.postgresql.org/docs/current/auth-pg-hba-conf.html)
  and the `local auth-method` and why a password is required when connecting from the `api` container, but not the `db` container

- You learned about `migrations` and their use case

  Later, when we get to Kubernetes, we'll learn how to ensure that migrations
  are run prior to auto-booting the server.

- The way we are intializing a database conection it in our `api` is not written
  well in its current state, because the code is not quite testable. We will
  expand on what that means and how to fix that with encapsulation in a later
  section

## Fix Broken Imports in VSCode

Open [./api/main.go](./api/main.go).

If you are in VSCode and your editor is complaining about broken imports, first
be sure to install the golang dependencies on your workstation (outside of
the container).

```
$ cd api
$ go mod download
```

Then restart your golang language server
`(CMD + SHIFT + P > Go: Restart Language Server)` (CTRL on Windows).

> Note: you can may be able away with mounting your GOROOT and GOPATH to your
> `api` container, but this comes with implications such as your workstation
> architecture not matching your container's, thus binaries installed via go
> will not work.

The editor should stop complaining, and the `Go to Definition` feature should
now work.

## Adding CRUD Features to API

In this section, we'll focus on core functionality for our blog posts. While
not required, we will use an object relational mapping tool (`ORM`) to drive
a lot of the database interactions from our `api`. The purpose of using an ORM
is to stray way from writing as much SQL as we can. We will use [gorm](https://gorm.io/docs/).

### Create a Blog Record

In your `api` container, add the `gorm` dependency and associated postgresql
driver:

```
$ go get -u gorm.io/gorm
$ go get -u gorm.io/driver/postgres
```

> Note: you might notice that `gorm` supports a form of migrations. It does
> not seem to support migrations at the file-based level, but instead only
> "auto-migrations", which can make it difficult to trace changes to the
> database schema over time. We will not use that feature, but must take note
> that the type definitions for our models must conform to our migrations,
> which are now loosely coupled.

Next, let's create a `type` for our blog model. This will be used to marshal
an incoming JSON body to our model. Create the following directory and
file on your workstation:

```bash
$ mkdir api/models
$ touch api/models/blog.go
```

> Note: read more about declaring [gorm models](https://gorm.io/docs/models.html).
> We are using `gorm.Model` to auto-include fields from `gorm`, as well as
> additional utility from the ORM itself. This also includes fields, like
> `created_at`, `updated_at`, and `deleted_at`. Remember, because we are using
> migrations external to `gorm`'s auto-migrate feature, we need to explicitly
> specify those fields in our migration files. This is especially important
> if we want to leverage [associations between models using gorm](https://gorm.io/docs/belongs_to.html).

```
package models

import (
	"gorm.io/gorm"
)

type Blog struct {
	gorm.Model
	Title string `json:"title" binding:"required"`
	Body  string `json:"body" binding:"required"`
}

```

Below, we configure `main.go` to:

1. Connect to our postgres server using the connection string `POSTGRESQL_URL`  (propagated by our container via .env and docker compose)
2. Enable logging of SQL statements executed by `gorm`
3. Add a new HTTP endpoint responsible for creating a blog record

```diff
package main

import (
	"net/http"
+	"os"

+	models "example.com/m/v2/models"
	"github.com/gin-gonic/gin"

+	"gorm.io/driver/postgres"
+	"gorm.io/gorm"
)

func main() {
+	db, err := gorm.Open(postgres.Open(os.Getenv("POSTGRESQL_URL")), &gorm.Config{
+		Logger: logger.Default.LogMode(logger.Info),
+	})
+	if err != nil {
+		panic("failed to connect database")
+	}

	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
+	r.POST("/blogs", func(c *gin.Context) {
+		var blog models.Blog
+		c.BindJSON(&blog)
+		db.Create(&models.Blog{Title: blog.Title, Body: blog.Body})
+	})
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
```

At this point, we should be able to create a blog record.

Start the server in the `api` container:

```
$ go run main.go
```

Test the creation from your workstation:

```bash
$ curl -X POST http://localhost:8080/blogs \
   -H 'Content-Type: application/json' \
   -d '{"title":"my first blog","body":"hello world!"}'

$ echo $?
# 0
```

If curl does not complain with a non-zero exit code, then that likely means you
received an HTTP 200 (or 200 range) response code. Check the log output from the
api server:

```
# Log output from go run main.go
[GIN] 2023/05/18 - 07:22:10 | 200 |      1.2798ms |      172.26.0.1 | POST     "/blogs"
```

Looking good so far. Let's verify in the `db` container that there is a blog
record created:

```
$ docker compose exec db bash
root@bb9a81cc65d6:/# psql -U postgres
psql (15.3 (Debian 15.3-1.pgdg110+1))
Type "help" for help.

postgres=# select * from blogs;

 blog_id |     title      |     body
---------+----------------+--------------
       1 | my first blog  | hello world!
```

Yay! Our record has been persisted. What great news!

Key takeaways:

- ORM tools can help expedite development, but come with overhead, such as
  learning curves, especially if you need to write complex queries
- In a real application, you would want to put more effort into the behavior
  around HTTP response codes here

  For example, if you re-send that same curl command, you'll see the following
  output from the server console:

    ```
    2023/05/18 07:47:06 /src/main.go:29 ERROR: duplicate key value violates unique constraint "blogs_title_key" (SQLSTATE 23505)
    [0.643ms] [rows:0] INSERT INTO "blogs" ("title","body") VALUES ('my first blog','hello world!')
    ```

  You would want to account for such an error, and return appropriate HTTP
  status code to the client, such as a `409 conflict`. We'll make this change
  later when we improve the code and make it more testable.

- The way that our `gin` routes are implemented is not exactly ideal,
  what if we needed to handle some complex before or after this point in the
  application? For example, if we needed to manipulate or impose checks on
  our models or incoming data, we could see that could get ugly real fast. We
  will address this in a later section

### Get All Blogs

For the next few sections, you can refer to the [gorm query documentation](https://gorm.io/docs/query.html)
for how to query your database. In this section, we'll:

- Implement an endpoint for fetching all blog records
- Create a second record
- Manually test the new endpoint

First, let's implement the endpoint `/blogs`. If an `HTTP GET request` is sent
to this route, return all blogs. Let's edit `main.go`:

```diff
package main

import (
	"net/http"
	"os"

	models "example.com/m/v2/models"
	"github.com/gin-gonic/gin"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	db, err := gorm.Open(postgres.Open(os.Getenv("POSTGRESQL_URL")), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		panic("failed to connect database")
	}

	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
+	r.GET("/blogs", func(c *gin.Context) {
+		var blogs []models.Blog
+		db.Find(&blogs)
+		c.JSON(http.StatusOK, blogs)
+	})
	r.POST("/blogs", func(c *gin.Context) {
		var blog models.Blog
		c.BindJSON(&blog)
		db.Create(&models.Blog{Title: blog.Title, Body: blog.Body})
	})
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

```

Next, let's restart the `api` server by hitting `ctrl+c` in the terminal that is
running `go run main.go`, and rerunning that command. Now create a second record
by running the following from your workstation:

```
$ curl -X POST http://localhost:8080/blogs \
   -H 'Content-Type: application/json' \
   -d '{"title":"my second blog","body":"hello world!"}'
```

And then query our `api` for all of the blogs created so far:

```bash
$ curl http://localhost:8080/blogs

# Output
[
  {
    "ID": 1,
    "CreatedAt": "2023-05-19T04:01:30.113952Z",
    "UpdatedAt": "2023-05-19T04:01:30.113952Z",
    "DeletedAt": null,
    "title": "my first blog",
    "body": "hello world!"
  },
  {
    "ID": 2,
    "CreatedAt": "2023-05-19T04:49:48.037571Z",
    "UpdatedAt": "2023-05-19T04:49:48.037571Z",
    "DeletedAt": null,
    "title": "my second blog",
    "body": "hello world!"
  }
]
```

### Get a Single Blog

It is common in REST to fetch records by some field. In this case, we'll use
the ID field--you could even use the blog title if you wanted to. When we
send an `HTTP GET` request to `/blogs/:id`, it should return the blog with that
ID.

In `main.go`, add the following route:

```go
r.GET("/blogs/:id", func(c *gin.Context) {
  id := c.Params.ByName("id")
  fmt.Println("id qqq", id)

  var blog models.Blog
  db.Find(&blog, id)
  c.JSON(http.StatusOK, blog)
	})
```

Restart the `api` server and run the following from your workstation:

```bash
$ curl http://localhost:8080/blogs/1
{
  "ID": 1,
  "CreatedAt": "2023-05-19T04:01:30.113952Z",
  "UpdatedAt": "2023-05-19T04:01:30.113952Z",
  "DeletedAt": null,
  "title": "my first blog",
  "body": "hello world!"
}


$ curl http://localhost:8080/blogs/1
{
  "ID": 2,
  "CreatedAt": "2023-05-19T04:49:48.037571Z",
  "UpdatedAt": "2023-05-19T04:49:48.037571Z",
  "DeletedAt": null,
  "title": "my second blog",
  "body": "hello world!"
}
```

### Update a Blog

In this section, we'll give users the ability to replace an existing blog. This
would be similar to editing it, but in-code, we must replace the entire record
(instead of a single field). It is typically easier to do this than to implement
the ability to replace only a single field at a time (by which we would use
an `HTTP PATCH` request). This replacement method we are using will use an
`HTTP PUT` request to `/blogs/:id`. It should take some blog input, and replace
it in the database.

In `main.go`, add the following route:

```go
r.PUT("/blogs/:id", func(c *gin.Context) {
  id, err := strconv.ParseUint(c.Params.ByName("id"), 10, 64)
  if err != nil {
    fmt.Println(err)
  }

  var blog models.Blog
  c.BindJSON(&blog)

  blog.ID = uint(id)

  db.Save(&blog)
  c.JSON(http.StatusOK, blog)
})
```

Restart the `api` server. Now let's replace the first blog by running the
following command on your workstation:

```bash
$ curl -X PUT http://localhost:8080/blogs/1 \
   -H 'Content-Type: application/json' \
   -d '{"title":"my first edited blog","body":"hello world!"}'

# Output
{
  "ID": 1,
  "CreatedAt": "0001-01-01T00:00:00Z",
  "UpdatedAt": "2023-05-19T05:58:09.8155476Z",
  "DeletedAt": null,
  "title": "my first edited blog",
  "body": "hello world!"
}
```

And you can confirm it is edited by fetching it again and confirming its output
is expected.

```bash
$ curl http://localhost:8080/blogs/1
```

Key takeaways:

- You can see that there is some typecasting happening here. If an error were
  to occur for whatever reason, we are not handling that adequately. We will
  address this in a future section

### Delete a Blog

In this section, we'll give users the ability to delete a blog. This is done
by sending an `HTTP DELETE` request using an identifier for the resource we're
targeting--in this case, the id. So to delete our first blog, we'd send the
request to `/blogs/1`.

In `main.go`, add the following:

```go
r.DELETE("/blogs/:id", func(c *gin.Context) {
  id := c.Params.ByName("id")
  db.Delete(&models.Blog{}, id)
  c.JSON(http.StatusOK, nil)
})
```

Now let's delete our first blog:

```bash
$ curl -X DELETE http://localhost:8080/blogs/1
```

You should not receive any output, but see in the `api` server console, the
following was logged, indicating the deletion went through:

```
2023/05/19 06:19:33 /src/main.go:64
[2.142ms] [rows:1] UPDATE "blogs" SET "deleted_at"='2023-05-19 06:19:33.261' WHERE "blogs"."id" = '1' AND "blogs"."deleted_at" IS NULL
[GIN] 2023/05/19 - 06:19:33 | 200 |      2.2975ms |      172.26.0.1 | DELETE   "/blogs/1"
```

Try to fetch it to confirm deletion:

```bash
$ curl http://localhost:8080/blogs/1
```

You should not receive any output, but notice the API server still reports
it sent an `HTTP 200` status code.

```
2023/05/19 06:19:59 /src/main.go:40
[0.533ms] [rows:0] SELECT * FROM "blogs" WHERE "blogs"."id" = '1' AND "blogs"."deleted_at" IS NULL
[GIN] 2023/05/19 - 06:19:59 | 200 |       671.5µs |      172.26.0.1 | GET      "/blogs/1"
```

That is not ideal, and should be an `HTTP 404`. As you can see, there are
several things that our application needs improvement upon.

Key takeaways:

- Fetching a non-existing record returns a 200 instead of a 400 HTTP status code
- Anyone can CRUD our blog resources, which is not very secure

These will be addressed in a future section.

## Setup the Go Debugger

We will use [delve](https://github.com/go-delve/delve/tree/master/Documentation/installation) to enable us to step
through code, make debugging easier. This will enable us to step through code
without having to resort to print statements...

Add the following to [api/Dockerfile](./api/Dockerfile):

```dockerfile
# Install golang debugger
RUN go install github.com/go-delve/delve/cmd/dlv@latest
```

Also run the above `go` command on your workstation, as we will need to use
`dlv` in client mode to connect to the debugging server.

Also expose an arbitrary port that we will use for the debugging server. Expose
that port in [docker-compose.yml](./docker-compose.yml).

```diff
ports:
  - "8080:8080"
+  - "4000:4000"
```

Rename [Dockerfile](./api/Dockerfile) to [Dockerfile.dev](Dockerfile.dev).
In the new file, change the `WORKDIR` path to `api`. This is because
`delve` and `vscode` maintain the path of the workspace directory when
communicating to the debug server.

```diff
# Where our application will live in the completed container
-WORKDIR /src
+WORKDIR /api
```

Create a `.vscode/launch.json` file with the following contents. This
configuration will allow you to connect to the `delve` debugger server
once we run it:

> Note: the use of [substitutePath](https://github.com/golang/vscode-go/blob/master/docs/debugging.md)
> will replace the path of your files up to the root of this repo with an empty
> string. This ensures that the debugging server can find the files that you
> mark with a breakpoint relative to the server itself (in-container).
>
> For example, the path `/home/$USER/git/repo/api/main.go` would appear
> as `api/main.go` in the docker container. The debugging server will
> be able to find this relatively due to the volume mounts we've
> defined in [docker-compose.yml](./docker-compose.yml).

```json
{
  "version": "0.2.0",
  "configurations": [
    {
        "name": "Remote API Server",
        "type": "go",
        "request": "attach",
        "mode": "remote",
        "port": 4000,
        "host": "127.0.0.1",
        "showLog": true,
        "trace": "verbose",
        "substitutePath": [{
          "from": "${workspaceFolder}",
          "to": ""
        }],
    }
  ]
}
```

Stop your compose project, and restart it, ensuring that the `api` container
is rebuilt:

```bash
# if not detached, ctrl+c
# if detached, run
$ docker compose down
$ docker compose up --build
```

Test the delve command in the `api` container:

```
$ docker compose exec api bash
root@3da06136987d:/src# dlv version
Delve Debugger
Version: 1.20.2
Build: $Id: e0c278ad8e0126a312b553b8e171e81bcbd37f60
```

Create a breakpoint on any line in your `api` codebase.

Run the delve server by running the script below from your workstation:

```bash
./bin/debug --start
```

Connect to the debugging server in vscode using the `Remote API Server`
launch configuration profile that we created.

Start debugging on the client side in your IDE (using `F5` in vscode).

If your breakpoint is within a route, be sure to query your webserver.
When your breakpoint is hit, your vscode editor should display the
debugging information and allow you to step through the code.

> Note: currently, I could not figure out how to get vscode and delve
> to play nicely when stepping through a 3rd party go library/module.
> The vscode client attempts to open those under your vscode
> [${workspaceFolder}](https://code.visualstudio.com/docs/editor/variables-reference)
> with the `GOROOT` concatenated on the end.

To stop the debug server, disconnect from the server from your IDE.

> Note: if you started the debug server from inside the container with the
> above script, you can stop it by running this from within the container:
> `./bin/debug --stop`.

### Running Tests

## Testing I

Testing can be very complicated if the codebase is poorly written and if you
do not have a test plan. This is the state that this codebase is currently in.
We will focus on unit, integration, and end-to-end (E2E) tests with the
following goals/definitions of each of them:

- Unit: a single functionality. A function, class, or module that does one
  thing. You are testing that single code path

    Oftentimes, you'll be testing the implementation of this piece of code.
    This includes errors and parameter permutations. We want to be careful
    of testing too many permutations at a higher level, as this can increase
    the time our tests take substantially.

- Integration: a combination of modules or services. You are testing a code path
  between public faces interfaces in-code or services
- E2E: a fully functioning application and related services. You are testing
  how the system behaves from the end-user perspective

Test writing typically follows the Four Phase test plan:

- Setup
- Exercise
- Verify
- Teardown

One of the boons of golang is that it is not Object Oriented.
Relationships are always established through interfaces, as opposed to
inheritence. This generally means that we can have a `test double` for any
behavior so long as the interface fits.

A `test double` has many aliases, as some tests might have
different goals, achieved with `mocks`, `fakes`, and `spies`
(which will be explained more later), but in short, it allows us to control the
flow of our code in tests to achieve a specific testing goal. Using test doubles
makes unit testing (and in some cases, integration testing) simpler, faster,
and more targeted. However, we need to be weary of overtesting, which mocking
can lead to if we are not testing intelligently. We should always ensure that
we are testing *some meaningful behavior*. We'll decipher what that means in
some examples below.

You will often hear the term `mocking` used interchangably with one of the
many flavors of `test doubles`.

As an aside, in Object Oriented Programming (`OOP`), unit testing can also
be made easier if you have an abstract class that many sub-classes inherit from,
given the inherited methods are not overriden, you would need fewer tests to
cover that behavior as you only need to test the base method as needed. As
always, this still subject to how well written the code is.

Go provides an in-built testing framework
[(go test)](https://pkg.go.dev/testing). We will be leveraging that.

### A Handful of Unit Tests

First, we don't really have anything to test at the unit level
because we are leveraging so much third-party tooling that is well-tested
already. We should aim to test core functionality of our business logic.

Let's create a function `GetWordCount` on the `blog` model that counts the
number of occurances of words in the a blog post.

```go
# api/models/blog.go
func (b *Blog) GetWordCount() map[string]int {
	m := make(map[string]int)

	words := strings.Split(b.Body, " ")

	// For each word
	for _, element := range words {
		// Check if in map
		_, ok := m[element]
		if ok {
			// If so, increment
			m[element] += 1
		} else {
			// Otherwise, init to 1
			m[element] = 1
		}
	}

	return m
}
```

Let's add an endpoint to serve this functionality:

```go
// main.go
r.GET("/blogs/:id/words", func(c *gin.Context) {
  id := c.Params.ByName("id")

  var blog models.Blog
  db.Find(&blog, id)
  wc := blog.GetWordCount()

  c.JSON(http.StatusOK, wc)
})
```

> Note: be careful not to abuse REST Principles by turning your API into a
> series of Remote Procedure Calls. One could argue that this data is or is not
> a valid "resource". Since this is not modifying state, and access is driven
> by HTTP Verbs, we'll allow it for now. The purpose here is to really give
> you an interface to invoke this code (perhaps in conjunction with the
> debugger) to see its output before we introduce you to the testing framework.

Next, let's rerun the server, then count words in our second blog post:

```bash
$ curl localhost:8080/blogs/2/words
{"hello":1,"world!":1}
```

Let's create another blog post to test:

```bash
$ curl -X POST http://localhost:8080/blogs \
   -H 'Content-Type: application/json' \
   -d '{"title":"my fourth blog","body":"red red red blue green green yellow yellow yellow yellow"}'

$ curl localhost:8080/blogs/4/words
{"blue":1,"green":2,"red":3,"yellow":4}
```

Perfect! Now we have some of our own business logic to test.

Let's create `api/models/blog_test.go` to confirm this behavior:

```bash
package models

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWordCount(t *testing.T) {
	var blog = &Blog{
		Title: "test title",
		Body:  "red red red blue green green yellow yellow yellow yellow",
	}

	result := blog.GetWordCount()
	expected := map[string]int{
		"blue":   1,
		"green":  2,
		"red":    3,
		"yellow": 4,
	}

	assert.True(t, reflect.DeepEqual(result, expected), "The two word counts be the same.")
}
```

Run the test from the `api` directory on your workstation:

```bash
# This command runs all tests recursively
$ go test ./...
```

Or, run it directly from the `api/models` directory:

```bash
# This command only tests files within the current directory.
$ go test
PASS
ok      example.com/m/v2/models 0.003s

# This command shows more details on what tests were run.
$ go test -v
=== RUN   TestWordCount
--- PASS: TestWordCount (0.00s)
PASS
ok      example.com/m/v2/models 0.003s
```

Great! Let's create a dependency that we can use in an integration test later.
First, we'll unit test that new dependency. From the `api` directory in
your workstation, create the following folders and files:

```bash
$ mkdir -p pkg/utils
$ touch strings.go
$ touch strings_test.go
```

Populate `strings.go` with the following code. We will use this to strip out
characters in a given string:

```go
package strings

import (
	"regexp"
)

// Returns a string with all non-word characters removed.
func ReplaceSymbols(s string) string {
	m := regexp.MustCompile("[^a-zA-Z0-9]")
	return m.ReplaceAllString(s, "")
}

```

In the `strings_test.go`, let's test a few different input scenarios:

```go
package strings

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReplaceSymbols(t *testing.T) {
	s := "h!e.l/l>o$w/o\\r,l<d"

	assert.Equal(t, "helloworld", ReplaceSymbols(s), "All symbols are replaced")
}

func TestReplaceSpace(t *testing.T) {
	s := "hello world"

	assert.Equal(t, "helloworld", ReplaceSymbols(s), "A single space is replaced")
}

func TestReplaceMultipleSpaces(t *testing.T) {
	s := "he  l  l  o w               o r l        d"

	assert.Equal(t, "helloworld", ReplaceSymbols(s), "Multiple spaces in sequence are replaced")
}

```

Now let's run these test by running the following commands from your `api`
container:

```
$ cd pkg/utils
$ go test -v
=== RUN   TestReplaceSymbols
--- PASS: TestReplaceSymbols (0.00s)
=== RUN   TestReplaceSpace
--- PASS: TestReplaceSpace (0.00s)
=== RUN   TestReplaceMultipleSpaces
--- PASS: TestReplaceMultipleSpaces (0.00s)
PASS
ok      example.com/m/v2/pkg/utils      0.003s
```

Perfect! Now we have two units that we will test at an integration level in
the next section.

### Unit Testing, Integration Testing, and Stubs I

Let's add requirements for the `GetWordCount` feature from a blog are to ignore
*all* non-alphanumeric symbols.So, `!blue blue` should result in two counts of
`blue`, and `re!d r$ed r><ed` should result in three counts of `red`. It just so
happens that we have two units that, when combined, can satisfy this
requirement.

In this section, we'll update `GetWordCount` to use the
`strings.ReplaceSymbols()` method that we created in the last section, and write
a test to verify that new behavior.

Let's update `models/blog.go`:

```diff
package models

import (
	"strings"

	"gorm.io/gorm"

	utils "example.com/m/v2/pkg/utils"
)

type Blog struct {
	gorm.Model
	Title string `json:"title" binding:"required"`
	Body  string `json:"body" binding:"required"`
}

func (b *Blog) GetWordCount() map[string]int {
	m := make(map[string]int)

	words := strings.Split(b.Body, " ")

	// For each word
	for _, word := range words {
		// Check if in map
-		_, ok := m[word]
+		_, ok := m[utils.ReplaceSymbols(word)]
		if ok {
			// If so, increment
			m[word] += 1
		} else {
			// Otherwise, init to 1
			m[word] = 1
		}
	}

	return m
}

```

Now `blog.GetWordCount()` has a dependency on our `utils` package. One could
argue that the problem with this code is that is not very testable, because
we cannot use a `test double` for the call to `utils.ReplaceSymbols(word)` due
to the way we've included it here. Therefore, we cannot test
`blog.GetWordCount()` at the unit level, we will have to test it at the
integration level.

If we could test this at the unit level, we would want to `stub` the call
to `utils.ReplaceSymbols(word)`. A `stub` is a type of `test double` that simply
enforces some arbitrary return value for the sake of testing only. No
functionality is tested in the code that we are stubbing out itself,
the hard-coded response is used to drive the rest of the code down a specific
path for testing. Stubbing out that call might look like the following code:

```go
// blog.go
func (b *Blog) GetWordCount( replaceSymbols func) map[string]int {
	m := make(map[string]int)

	words := strings.Split(b.Body, " ")

	// For each word
	for _, word := range words {
		// Check if in map
-		_, ok := m[word]
+		_, ok := m[replaceSymbols(word)]
		if ok {
			// If so, increment
			m[word] += 1
		} else {
			// Otherwise, init to 1
			m[word] = 1
		}
	}

	return m
}

// blog_test.go
func replaceSymbolsStub(s string){
  // This is hard-coded, no functionality exists here
  return "blue blue"
}

func TestWordCountWithStub(t *testing.T) {
	var blog = &Blog{
		Title: "test title",
		Body:  "!blue blue",
	}

  // This call returns the hard-coded value "blue blue"
	result := blog.GetWordCount(replaceSymbolsStub)
  // Which gets processed intot the following...
	expected := map[string]int{
		"blue":   2,
	}

	assert.True(t, reflect.DeepEqual(result, expected), "The two word counts be the same.")
}

```

> Note: in reality, you might mock something out if you want to assume that
> will return a successful or erroneous result. In such scenarios, you are not
> testing the external functionality, but the *inteface* of the dependency in
> conjunction with our code and the code path that follows.

On the flip side, one could argue that you do not need to use a `stub` here.
The function `utils.ReplaceSymbols()` does not have any side effects. It does
not mutate state anywhere inside or outside of itself. Therefore, it might
be adequate to test that function at the unit level in isolation, and to forego
any `stubbing or mocking` of it here.

However, the lack of the ability to stub here means that we cannot write a unit
test for `blog.GetWordCount()`, but instead, tests around this method would be
considered an `integration test`. This is because because we have two components
`ReplaceSymbols()` and `GetWordCount()` that are working together, unstubbed.

Key takeaways:

- The nuance between unit and integration testing in this case is whether or not
  the external call to a separate component can be stubbed out
- Because the ability to stub out `utils.ReplaceSymbols()` in our code does not
  exist, it could be considered a code smell, because we cannot test
  `blog.GetWordCount()` at the unit level
- In order to test `blog.GetWordCount()` at the unit level, it would require our
  code structure to change (be it function signatures, and encapsulation or use
  of a `factory pattern` for the `utils` package)

  It is up to your discretion as to whether or not such attention to detail is
  required. One could argue that if our `utils` package remains stateless, that
  encapsulating it to enable the abilitry to `stub` out calls to it would be
  over engineering.

  Now, apply this logic to an API client that you've written to interact with
  another service. You would likely want that to be as testable as possible,
  because it would be core to the functionality of your product. The more
  code paths you test at the unit level, the fewer tests you might need at the
  higher level, ultimately saving you time in CI/CD pipelines and developer
  feedback loops.

### Unit Testing, Integration Testing, and Stubs II

In the last section, we changed our business logic in a way that prevents us
from testing a specific function at the unit level by introducing a
dependency that cannot be stubbed. This means that the lowest level test we
can write for this function is an integration test. That is arguably okay in
some scenarios, but let's explore scenarios where that might not be the case.

We mentioned earlier that the state of our code is not ideal at the moment. Take
a look at the following route and ask yourself how you would test this code?

```go
r.GET("/blogs/:id", func(c *gin.Context) {
  id := c.Params.ByName("id")

  var blog models.Blog
  db.Find(&blog, id)
  c.JSON(http.StatusOK, blog)
})
```

Here are some things that come to mind:

- The second parameter to `r.GET()` is an anonymous function--this cannot be
  tested at the unit or integration level

  However, it can be tested at the E2E level (to some degree), since I could
  start my server, query that endpoint, and verify its output.

- What if I wanted to verify that `db.Find()` actually found a record at the
  unit or integration level?

  Realistically, when testing at the unit level, it would be worth verifying
  if `db.Find()` was called, but not test the underlying the code of the 3rd
  party module. That would tell us that this interface is being used at the
  unit level, and receives the expected parameters. We would also `stub` out
  that method to simply mutate the given reference of `models.Blog` whoese value
  we would verify after the test. So the goal here--at the unit level--would be
  to verify this interface is being appropriately used.
  
  The difference betweenthis unit test, and the ones we wrote earlier is with
  this test, we check that the unit iscorrectly implemented and adhere to the
  expected interface contracts. Before, we were just verifying actual business
  logic and/or rules.

  This also implies that any error handling you implement on your own would need
  to be tested at the unit level, too. What happens if nothing is found and
  `blog` remains empty?

  When testing at the integration level, we may or may not have our database
  running during our test. Without mocking, this test would tell us that this
  interface is being used by not only the units we wrote tests for earlier, but
  also the thing that was stubbed out previously.

- What if I wanted to verify that the response body contained the appropriate
  values? How would I test that at the unit or integration level?

  First, you'd do a similar unit test as mentioned earlier, is confirm that
  the interface being used `c.JSON()` is correct. That way, if in the future,
  a developer changes the response body to use a different interface, this
  test would fail notifying the developer of a required code change needing
  to be made, or they made a mistake themselves. TODO: VERIFY.

  There is a design pattern called `middleware` used in many modern day web
  frameworks that allow you to execute something before and/or after a specific
  HTTP handler is called (the anonymous function we currently have for this
  route). We could test two separate HTTP handlers with the same
  `c *gin.Context` where we can expect and validate one of these middleware
  functions to manipulate our context in a specific way. That would be an
  example of a valid integration test.

With all of that said, you can see that this code is not testable mostly due
to the use of the anonymous function. The simple act of encapsulating
like-functionality into their own modules or packages makes interfacing with
them easier... and where there are interfaces, there is the capabiltiy to
double/stub/mock!

# A Major Refactor - Business and Software Architecture

I didn't intend for this to turn into a sort of tour of software design
patterns and the nuances between business and software architecture, but it felt
like a natural path to take when discussing the testability of code. So,
here we are!

We will cover these topics in this section:

- Enterprise Architecture (think: business)
- Onion Architecture (think: code)
- SOLID Principles (more code)

These are not mutually exclusive, and can actually compliment each other.
The reason I'm introducing you to them at the same time is so you can see
how they can enable each other.

Before you continue, know that the context here starts from a very high level
(in terms of software architecture and design), and gets incresingly lower
level. Everything that is discussed here aids and enables the other sections.
Imagine that we are looking at a big picture and slowly zooming in on the finer
details of the image.

## Enterprise Architecture

From Wikipedia:

> Enterprise architecture (EA) is a business function concerned with the 
> structures and behaviors of a business, especially business roles and
> processes that create and use business data. The international definition 
> according to the Federation of Enterprise Architecture Professional
> Organizations is "a well-defined practice for conducting enterprise analysis, 
> design, planning, and implementation, using a comprehensive approach at all
> times, for the successful development and execution of strategy. Enterprise
> architecture applies architecture principles and practices to guide 
> organizations through the business, information, process, and technology
>  changes necessary to execute their strategies. These practices utilize the 
> various aspects of an enterprise to identify, motivate, and achieve these
> changes."

From ChatGPT:

Enterprise architecture is a software architecture pattern that focuses on the
design and structure of large-scale enterprise-level systems. It provides a
holistic and strategic approach to managing and aligning an organization's IT
infrastructure, systems, applications, and processes with its business goals
and objectives. The goal of enterprise architecture is to enable efficient and
effective operation of an organization by facilitating the integration,
interoperability, and scalability of its IT systems.

At the code level, the principles of enterprise architecture translate into
certain practices and patterns that promote modularity, scalability,
reusability, and maintainability of the software systems within an organization.
Here are some key aspects of enterprise architecture at the code level:

Modular Design: Code is organized into modular components that encapsulate
specific functionalities or features. Modules can be implemented as libraries,
packages, or microservices. This modular design allows for better separation of
concerns and promotes code reuse and maintainability.

Service-Oriented Architecture (SOA): Systems are designed based on the
principles of service-oriented architecture. Functionality is encapsulated
as services that provide well-defined interfaces and can be independently
developed, deployed, and scaled. Services communicate with each other through
standardized protocols (e.g., REST, SOAP) and can be composed to build complex
applications.

API Design: Emphasis is placed on designing clean, well-documented APIs
(Application Programming Interfaces) that enable interoperability and ease of
integration between different systems and components. APIs provide a contract
for communication and ensure consistency and compatibility.

Standardization and Frameworks: Organizations establish coding standards, best
practices, and frameworks that promote consistency, readability, and
maintainability of the codebase. Standardized frameworks and libraries are used
for common functionalities such as authentication, logging, database access,
and error handling.

Scalability and Performance: Code is designed with scalability and performance
considerations in mind. Techniques such as caching, load balancing, asynchronous
processing, and distributed computing are utilized to ensure the system can
handle increased load and provide optimal performance.

Data Management: Code follows best practices for data management, including
proper database design, data access patterns (e.g., ORM frameworks), and data
consistency and integrity mechanisms. Enterprise architecture may include
guidelines for data governance, security, and compliance.

Error Handling and Logging: Robust error handling and logging mechanisms are
implemented to capture and handle errors effectively. Logging helps in
troubleshooting, monitoring, and analyzing system behavior.

Testing and Quality Assurance: Enterprise architecture emphasizes the importance
of testing and quality assurance practices. This includes unit testing,
integration testing, and other testing methodologies to ensure the reliability
and correctness of the codebase.

Continuous Integration and Deployment (CI/CD): Automation is applied to build,
test, and deploy code changes. CI/CD pipelines are set up to ensure rapid and
reliable delivery of software updates while maintaining quality and stability.

So far, we can see that we have touched on a couple of these things already,
all of which can be improved upon:

- Modular Design

    We have separated out some of our code, but we can do better
    in this regard.

- Service Oriented Architecture

    We are focused on following REST principles for our API. We don't have
    a lot of systems in this project (yet), but we have taken the first step
    in building an API that can be consumed by other components, such as a
    Web GUI, CLI, SDK, or another API, etc.

- Standardization and Frameworks
  
    We are using existing open source tooling as opposed to rolling our own.
    Sometimes it can't be helped that we have to build our own tools, but other
    times, it is a calculated risk to depend on 3rd party software.

- Data Management

  Currently our data access pattern is driven by an ORM.

- Testing and Quality Assurance

    We've done a little bit on this, but are looking to improve this by
    adhering to an architectural pattern.

Key takeaways:

- When designing software, we should be cognizant of how it will be used across
  the business

- Whether the business is large or small plays a factor in which shortcuts we
  can take

    We want to avoid over engineering our product as to not inhibit velocity,
    as well to avoid complexity where it might not be necessary. Awareness
    is important, especially if our product has the potential for integrating
    with other services, or expanding the number of services offered.

    Thankfully, this application will be very simple, but after the first
    testing section and the context from this section, we are beginning to
    understand why we need to improve our code and the implications that has
    at the business level and the software level.

## Onion Architecture

At the end of the first testing section, we discussed ways to improve our code
for the sake of testability. This section will introduce some deisgn patterns
that enable us to do so. Let's meet the Onion Architecture (AKA Hexagonal
Architecture).

![Image - Onion Architecture](./docs/onion-architecture.webp)

This pattern enforces the Dependency Inversion Principle (`DIP`), meaning
high-level modules should depend on abstractions rather than concrete
implementations. This should immediately make you think about `interfaces`.

This architecture is often rendered in several different ways, but we will
focus on the following layers:

1. Domain Layer

   Our `model` code lives in our domain layer. All business logic should be
   executed here. State mutation and manipulation of our data should exist
   here. All domain problem solving exists here.

   The domain layer is unaware of other layers, it only exposes public methods
   for use by higher level layers.

2. Application Service Layer

    The application service layer is designed to serve a client of some sort.
    It is an interface to the domain layer in that the service exposes an
    "operation" to a client, and the client can execute it. That operation
    invokes code within the domain layer. That is, the application service layer
    integrates and consumes the domain layer.
    
    We are writing a REST API, which is served via the HTTP protocol. So it
    might make sense for us to have a layer between the application service
    layer and the infrastructure layer, which we will call the controller layer.
    This controller layer would consume the operations exposed by the service
    layer.

    1. Controller Layer

        This layer recieves an HTTP request, and invokes our application service
        layer to complete the client's request. This gives us separation such
        that we can do things specific to the HTTP client

    If we were writing a server-side CLI within the same codebase that would
    be deployed within this docker container, you could think of it similarly
    like this "Controller Layer", as it could also interface with this
    appliaction service layer.
    
    For example, perhaps both the CLI and the REST API might be able to
    execute the same operations--Boom, code reuse.

    > Important: do not confuse this server-side CLI with a client-side CLI.
    > The client-side CLI would exist in the presentation layer, and would
    > communicate with our HTTP interface at the presentation layer.

3. Infrastructure Layer

    This code does not solve any domain problem. You can think of this as the
    code that interacts with other services, such as our CRUD operations.
    This is not an HTTP controller or server-side CLI that invokes the CRUD
    operations, but instead, this is the actual code that queries the database
    or external services.

    You'll often see this code refered to as the `Repository Pattern`.

    Operations exposed by this code is consumed by the service in order to
    complete operations.

4. Application / External Dependencies Layer

    This use the user-facing application. It might be from a web interface
    or a client-side CLI, or even an SDK. It could also be external services
    that observe the application state (like log tracing). Essentially, anything
    client-side.

This is what our testing strategy might look like when adhering to this
architectural pattern.

![Testing Strategy - Onion Model](./docs/testing-strategy.webp)

We will adapt our code to this pattern so that we can leverage the testing
strategy shown. We could dive into this right now, but I think this last section
will highlight some design principles that will enable us to do this well.

Before we do, let's add cover one final bit of theory, as we continue our
descending to the lower level.

## SOLID Principals

SOLID is a collection of five principles of software design, and is a topic
introduced by Robert Martin in his book
*Agile Software Development, Principles, Patterns, and Practices.*. There is an
excellent reference at the end of this section that has code examples for
each principle.

TODO: include code examples here?

1. Single Responsibility Principle (`SRP`)

    > A class should have one, and only one, reason to change.
    > –Robert C Martin

    Code with the fewest responsibilities is least likely to change. This
    is important when your code depends on code that might change outside
    of your control.

    Since Go is not an `OOP` language, we are left only with `interfaces` and
    `packages`. Interfaces lend software to composition (over say, inheritence).
    Packages are a means of collecting like functionality into a single
    application. In Go, it might be appropriate to think that a package should
    accomplish one thing, and one thing well. Packages that have many
    responsibilities are subject to change without cause. Code that follows
    `SRP` should be cohesive (with other code) but loosely coupled
    (from other code).

    An immediate codesmell in our codebase is the `pkg/utils` package. It
    might be more adequate to move our `strings.go` utilities under a
    `pkg/utils/strings` package so that it is obvious at a higher level what
    the package is designed to accomplish. Perhaps even drop the term `utils`
    entirely from the package name.

    Now, I'd argue that a package can still be well designed if it introduces
    many types and functionality around each are well isolated within the
    package (`models`). It makes sense to keep these consolidated, because
    we don't want to have to import a package for each model, that just seems
    cumbersome in practice.

    The next immediate codesmell is the anonymous HTTP handlers that we
    discussed in detail in the first testing section. A combination of SOLID
    principles and the Onion Arcitecture will improve this code
    dramatically.

2. Open / Closed Principle

    > Software entities should be open for extension, but closed for modification.
    > –Bertrand Meyer, Object-Oriented Software Construction

    Since code should not change, it should at least be open to extension.
    That can be done by embedding the original `type` that you want to change
    under a new type, and overriding the functionality you wish to change.

    Golang allows `types` shared within a package to access each others'
    private members. This means that in addition to our ability to override,
    we can also extend existing functionality by building upon it in our
    overrides.

3. Liskov Substitution Principle (`LSP`)

    The Liskov Substitution Principle states that objects of a superclass
    should be replaceable with objects of its subclasses without breaking the
    program.

    In other words, if a class A is a subtype of class B, then you should be
    able to use an object of class A wherever an object of class B is expected,
    and the program should still work correctly.

    Go doesn't support inheritance, only interfaces, and they are
    satisfied implicitly, rather than explicitly.

    The author of the article I reference at the end of this section has this
    to say about LSP:

    ```go
    // io.Reader
    type Reader interface {
            // Read reads up to len(buf) bytes into buf.
            Read(buf []byte) (n int, err error)
    }
    ```

    > *"Because io.Reader‘s deal with anything that can be expressed as a stream
    > of bytes, we can construct readers over just about anything; a constant
    > string, a byte array, standard in, a network stream, a gzip’d tar file,
    > the standard out of a command being executed remotely via ssh."*

    By adhering to the `LSP`, you can write code that is flexible and modular,
    allowing you to interchange different implementations of `io.Reader` without
    impacting the behavior of the code that consumes it. This promotes code
    reusability and enables you to work with various data sources seamlessly.

4. Interface Segregation Principle (`ISP`)

    > Clients should not be forced to depend on methods they do not use.
    > –Robert C. Martin

    These code snippets are from the second article I reference below:

    ```go
    // s is useless
    func addNumbers(a int, b int, s string) int {
      return a + b
    }
    ```

    Uselessness is less obvious when using structs:

    ```go
    type Database struct{ }
    func (d *Database) AddUser(s string) {...}
    func (d *Database) RemoveUser(s string) {...}

    // d.RemoveUsers() is useless to this function
    func NewUser(d *Database, firstName string, lastName string) {
      d.AddUser(firstName + lastName)
    }
    ```

    This usage of interfaces allows us to only specify what we need:

    ```go
    type DatabaseWriter interface {
      AddUser(string)
    }

    type Database struct{}

    func (d *Database) AddUser(s string) {
      // Implementation of AddUser for Database
      // Add user to the database
    }

    // NewUser creates a new user using the provided DatabaseWriter implementation
    func NewUser(d DatabaseWriter, firstName string, lastName string) {
      d.AddUser(firstName + lastName)
    }

    func main() {
      db := &Database{} // Create an instance of Database
      NewUser(db, "John", "Doe")
    }
    ```

    In this example, `NewUser` accepts any implementation of `DatabaseWriter`,
    including `Database`. When you call `NewUser` with an instance of `Database`,
    it will invoke the `AddUser` method implemented in the `Database` type.

    This allows you to use the `Database` type within the context of the
    `NewUser` function while satisfying the requirements of the `DatabaseWriter`
    interface.

    Dave Cheny's take on `LSP`:

    > The results has simultaneously been a function which is the most specific
    > in terms of its requirements–it only needs a thing that is writable–and
    > the most general in its function.


5. Dependency Inversion Principle (`DIP`)

    > High-level modules should not depend on low-level modules.
    > Both should depend on abstractions. Abstractions should not depend on
    > details. Details should depend on abstractions.
    > –Robert C. Martin

    Surprise, this principle is core to the Onion Architecture. `DIP` ultimately
    means that the dependency graph should be asyclic. With regards to go,
    this pertains to your packages.

    From Cheny:

    > All things being equal the import graph of a well designed Go program
    > should be a wide, and relatively flat, rather than tall and narrow. If
    > you have a package whose functions cannot operate without enlisting the
    > aid of another package, that is perhaps a sign that code is not well
    > factored along package boundaries.
    >
    > The dependency inversion principle encourages you to push the
    > responsibility for the specifics, as high as possible up the import
    > graph, to your main package or top level handler, leaving the lower
    > level code to deal with abstractions–interfaces.

Wow. With all of that said, I think it's time to put this knowledge to good use
and see how we can reap the benefits!

Key takeaways:

- While a lot has been covered, it should be clear by now that composing our
  application with as many clearly and well-defined units as possible at
  the lower level will empower us in vastly many ways at the highest level

  This will give us less error-prone, more reusable, and more testable code.
  The more testable the code is, the less time we have to spend testing it
  at a higher level. The more reusable and well written (exposed), the easier
  it becomes to integrate with our product. This is the start of writing
  sustainable, scalable code!

References

- [Dave Cheny - SOLID principles in Go](https://dave.cheney.net/2016/08/20/solid-go-design).
- [Jack Lindamood - What Accept Interfaces, Return Structs Means](https://medium.com/@cep21/what-accept-interfaces-return-structs-means-in-go-2fe879e25ee8)


### Unit Tests and Mocks

A mock actually verifies that the mocked out methods were called (a spy).

TODO

### A Simple Integration Test

Let's say the requirements for this feature are to ignore *all*
non-alphanumeric symbols.So, `!blue blue` should result in two counts of `blue`,
and `re!d red` should result in two counts of `red`. It just so happens that
we have two units that, when combined, can satisfy this requirement.

Let's 

TODO

### A Simple E2E Test

TODO
