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
  - [Prerequisites](#prerequisites-1)
  - [Initializing the API Application](#initializing-the-api-application)

## How to Use

This repository is provided as-is. It is recommened that you follow the
instructions below and execute them yourself in an empty directory, and use
this repository as a completed example.

> Pro tip: step through and review each commit in the git history.

## Topics

1. [Docker](#docker)
   1. [Docker Compose](#docker-compose)
   2. Composing a microservice application
   3. Writing tests in Go
   4. Using the golang debugger
2. Kubernetes

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

3. Kubernetes Operators
   
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

## Prerequisites

You should already be familiar with the following docker topics:

1. How to build your own docker container
2. How to run a docker container
3. Volume mounts
4. Environment variables
5. Docker networking (and DNS resolution)

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
