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

TODO


