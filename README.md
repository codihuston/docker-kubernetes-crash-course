# Purpose

This repository is a crash course on how to develop container-native
applications.

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

Key things to note:

- Dependencies for the developer

  In order to run our golang application, we had to install golang and packages
  associated with our application. In the real world, applications can become
  very complex, with a large number of dependencies and tooling. It can be
  difficult to get these things installed sometimes, which can lead to the
  "it works on my machine" meme. This is one of the key areas that docker
  improves quality of life.

In the next section, we'll move our project to a docker compose project from
which we will develop.
