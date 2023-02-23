# IAM Prototypes

This repository contains several implementations of an authorization library that takes key information about a user and the resource it attempts to access and validates whether this action is allowed or not.

These implementations are all in Go and their corresponding underlying authorization libraries are:

- [github.com/casbin/casbin](https://github.com/casbin/casbin)
- [github.com/mikespook/gorbac](https://github.com/mikespook/gorbac)
- [github.com/osohq/go-oso](https://github.com/osohq/go-oso)

## Project Structure

### Library Implementations

Implementations reside in [./auth](./auth) where each of them have their own package. These all these libraries implement the `AuthLib` interface at [./auth/auth.go](./auth/auth.go) and can be instantiated using the `New<Library>Auth()` functions in each of the packages.

### Library Tests

In each implementation, there is a `test` directory that contains unit tests for that library, showcasing their usages.

### Protobuf Files

In the [./pb](./pb) directory, you can find a sample `gRPC` server proto file with its message types and its corresponding Go output, compiled using [gogo/protobuf](https://github.com/gogo/protobuf)'s `protoc-gen-gofast` plugin.

### Service and Client

In [./service](./service/) and [./client](./client) directories, you can find a Go implementation of the gRPC server described in [Protobuf Files](#protobuf-files) section and an example client implementation that connects to that server.

The server has the capability to use each of these libraries seamlessly. The library choice is currently `hard-coded` 😢, but there will be an option to choose it later. The client has no idea which auth library is used.

## Running Server and Client

Run the server with:

```sh
go run service/daemon/main.go
```

The client is implemented as a unit test, so running that test will run the client code:

```sh
go test -run ^TestClientCreds$ github.com/umutozd/iam-prototypes/client
```
