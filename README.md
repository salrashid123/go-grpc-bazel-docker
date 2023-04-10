# Deterministic builds with go + bazel + grpc + docker

Deterministic container images for gRPC+golang [bazel](https://bazel.build/).

The following sample will build a golang gRPC client/server and then embed the binaries into container images.

These images are will have a consistent image hash no matter where it is built

*  `greeter_server@sha256:59240a9675e02b7a4c0c24f4d3346afcedd229b4c38f1e24bd2e37afc87e7aac`
*  `greeter_client@sha256:cb1fdcd482f3a5a5523a631182befbc6aa6b9d083a7d5ea44eaae2fd6336c4d1`

For reference, see:

- [Building deterministic Docker images with Bazel](https://blog.bazel.build/2015/07/28/docker_build.html)
- [Create Container images with Bazel](https://dev.to/schoren/create-container-images-with-bazel-47am)
- [rules_docker](https://github.com/bazelbuild/rules_docker)
- [Deterministic builds with nodejs + bazel + docker](https://github.com/salrashid123/nodejs-bazel-docker)
- [Deterministic container hashes and container signing using Cosign, Bazel and Google Cloud Build](https://github.com/salrashid123/cosign_bazel_cloud_build)
- [Deterministic container images with java and GCP APIs using bazel](https://github.com/salrashid123/java-bazel-docker)
- [Deterministic container images with python and GCP APIs using bazel](https://github.com/salrashid123/python-bazel-docker)
- [Deterministic container images with c++ and GCP APIs using bazel.](https://github.com/salrashid123/cpp-bazel-docker)
- [Deterministic builds with nodejs + bazel + docker](https://github.com/salrashid123/nodejs-bazel-docker)

To run this sample, you will need `bazel` installed (see [Cloud Shell](#cloud-shell) for an easy way to use `bazel`)

In the end, you'll end up with the same digests

* Server:

```bash
$ docker pull salrashid123/greeter_server:greeter_server_image
$ docker inspect salrashid123/greeter_server:greeter_server_image
```

### With bazel docker container

The easiest way here it to run bazel in docker using the provided image.

[i know,its weird but the only thing we're using docker here for is for bazel...the build still happens deterministically]

```bash
git clone https://github.com/salrashid123/go-grpc-bazel-docker.git
cd go-grpc-bazel-docker

$ docker version
    Client: Docker Engine - Community
    Version:           20.10.12
    Server: Docker Engine - Community
    Engine:
      Version:          20.10.2

# server
docker run \
  -e USER="$(id -u)" \
  -v `pwd`:/src/workspace \
  -v /tmp/build_output:/tmp/build_output \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -w /src/workspace \
  gcr.io/cloud-builders/bazel@sha256:4a19236baf0e5d663942c3947497e3f5b5356ae3dd6f97b1fae92897a97a11ad \
  --output_user_root=/tmp/build_output \
  run  --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 greeter_server:greeter_server_image

# client
docker run \
  -e USER="$(id -u)" \
  -v `pwd`:/src/workspace \
  -v /tmp/build_output:/tmp/build_output \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -w /src/workspace \
  gcr.io/cloud-builders/bazel@sha256:4a19236baf0e5d663942c3947497e3f5b5356ae3dd6f97b1fae92897a97a11ad \
  --output_user_root=/tmp/build_output \
  run  --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 greeter_client:greeter_client_image
```


### With Cloud Shell

If you have access to Google Cloud Platform account, you can use Cloud Shell to run `bazel` and save yourself an installation.

```bash
gcloud alpha cloud-shell ssh 

git clone https://github.com/salrashid123/go-grpc-bazel-docker.git
cd go-grpc-bazel-docker
```

Then within the shell, you should be able to `bazel version` to ensure it is installed.


### Build Image with Bazel

Declare go dependencies from `go.mod` into `repositories.bzl` using gazelle:

```
$ bazel --version
bazel 5.0.0

bazel run :gazelle -- update-repos -from_file=go.mod -prune=true -to_macro=repositories.bzl%go_repositories
```

Then build the client and server

```bash
bazel build --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 greeter_server:all
bazel run  --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 greeter_server:greeter_server_image

bazel build --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 greeter_client:all
bazel run  --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 greeter_client:greeter_client_image
```

Note, the `BUILD.bazel` files for the client and server targets is set for a specific arch and os. eg

```bazel
go_binary(
    name = "server",
    embed = [":go_default_library"],
    visibility = ["//visibility:public"],
    goos = "linux", 
    goarch = "amd64",    
)
```

### Check Image

The output of the commands above will yield 

```bash
$ docker images
    REPOSITORY             TAG                    IMAGE ID       CREATED        SIZE
    bazel/greeter_client   greeter_client_image   f77d61a81826   52 years ago   31.5MB
    bazel/greeter_server   greeter_server_image   67ccf97f9421   52 years ago   31.6MB
```

Inspect the image thats generated.  The hash we're after is actually `RepoTags` which we'll generate and show later, for now

### (optional) Run the gRPC Client/Server

(why not, you already built it)

- with docker

```
docker run -p 50051:50051 bazel/greeter_server:greeter_server_image --grpcport :50051
docker run --network="host" bazel/greeter_client:greeter_client_image --host localhost:50051 -skipHealthCheck 
```

with bazel

```
bazel run --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 greeter_server:server -- --grpcport :50051
bazel run --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 greeter_client:client -- --host localhost:50051 -skipHealthCheck
```


with go

You will first want to build the files, see corresponding steps above

Then in `go.mod`:

```
module main

go 1.17

require (
	github.com/golang/protobuf v1.4.3 // indirect
	github.com/google/uuid v1.3.0 // indirect
	golang.org/x/net v0.0.0-20220127200216-cd36cc0744dd // indirect
	golang.org/x/sys v0.0.0-20211216021012-1d35b9e2eb4e // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/genproto v0.0.0-20200526211855-cb27e3aa2013 // indirect
	google.golang.org/grpc v1.44.0 // indirect
	google.golang.org/protobuf v1.25.0 // indirect
	github.com/salrashid123/go-grpc-bazel-docker/echo v0.0.0
)

replace github.com/salrashid123/go-grpc-bazel-docker/echo => ./echo
```

then

```bash
go run greeter_server/main.go --grpcport :50051 

go run greeter_client/main.go \
  --host localhost:50051 \
  -skipHealthCheck
```

### Specify docker image

Specify a docker repo to by setting the `repository` command here. In the case below, its container registry `gcr.io/project_id`

```bazel
go_image(
    name = "go_image",
    embed = [":go_default_library"],
    importpath = "main",
    visibility = ["//visibility:private"],
    goos = "linux",
    goarch = "amd64",
)

container_image(
    name = "greeter_server_image",
    base = ":go_image",  
    ports=["50051"],
    # repository = "docker.io/salrashid123"
    # repository = "gcr.io/PROJECT_ID"      
)
```

on push to a repo

- `Server`
```bash
$ bazel build --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 greeter_server:all
$ bazel run  --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 greeter_server:greeter_server_image

$ docker push gcr.io/$PROJECT_ID/greeter_server:greeter_server_image

```

you'll see the hash we need...this is specific and intrinsic to the image.

On any other machine, generate the builds and inspect

```bash
$ docker pull salrashid123/greeter_server:greeter_server_image
$ docker inspect gcr.io/PROJECT_ID/greeter_server:greeter_server_image
```


### Cloud Build

You can use Cloud Build to create the image by using the `bazel` builder and specifying the repository path to export to.  In the sample below, the repository is set o google container registry:

```yaml
go_image(
    name = "go_image",
    embed = [":go_default_library"],
    importpath = "main",
    visibility = ["//visibility:private"],
    goos = "linux",
    goarch = "amd64",
)

container_image(
    name = "greeter_server_image",
    base = ":go_image",  
    ports=["50051"],
    # repository = "docker.io/salrashid123"
    repository = "gcr.io/PROJECT_ID"      
)
```

Note that `cloudbuild.yaml` specifies the base bazel version by hash too

```yaml
steps:
- name: gcr.io/cloud-builders/bazel@sha256:4a19236baf0e5d663942c3947497e3f5b5356ae3dd6f97b1fae92897a97a11ad
  id: build
  args: ['run', '--platforms=@io_bazel_rules_go//go/toolchain:linux_amd64', 'greeter_server:greeter_server_image']

- name: gcr.io/cloud-builders/docker
  id: tag
  args: ['tag', 'us-central1-docker.pkg.dev/builder-project/repo1/greeter_server:greeter_server_image', 'us-central1-docker.pkg.dev/$PROJECT_ID/repo1/greeter_server']
  waitFor: ['build']

- name: 'gcr.io/cloud-builders/docker'
  id: push
  args: ['push', 'us-central1-docker.pkg.dev/$PROJECT_ID/repo1/greeter_server']
  waitFor: ['tag']

options:
  machineType: 'N1_HIGHCPU_32'
```


```bash
$ bazel clean
$ gcloud builds submit --config=cloudbuild.yaml --machine-type=n1-highcpu-32
```


### Attesting base dependencies

The `WORKSPACE` and git dependencies are all known down to the specific version of bazel and base container image

- `WORKSPACE`

The base image used for the client and server is [distroless](https://github.com/GoogleContainerTools/distroless):

```
container_pull(
    name = "distroless_base",
    digest = "sha256:75f63d4edd703030d4312dc7528a349ca34d48bec7bd754652b2d47e5a0b7873",
    registry = "gcr.io",
    repository = "distroless/base",
)
```

### Using Pregenerated protopb and gazelle

The default bazel configuration in `echo/BUILD.bazel` compiles the proto files.  If you would rather use pregenerated proto files (eg, to [avoid conflicts](https://github.com/bazelbuild/rules_go/blob/master/proto/core.rst#avoiding-conflicts), you must do that outside of bazel and just specify a library)

`A)` Generate `proto.pb`:

```bash
/usr/local/bin/protoc -I ./echo  \
  --include_imports --include_source_info \
  --descriptor_set_out=echo/echo.proto.pb \
  --go_opt=paths=source_relative \
  --go_out=plugins=grpc:./echo/ echo/echo.proto
```

`B)` comment the local `replace` directives in `go.mod`:

```
module main

go 1.17

require (
	github.com/golang/protobuf v1.4.3 // indirect
	github.com/google/uuid v1.3.0 // indirect
	golang.org/x/net v0.0.0-20220127200216-cd36cc0744dd // indirect
	golang.org/x/sys v0.0.0-20211216021012-1d35b9e2eb4e // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/genproto v0.0.0-20200526211855-cb27e3aa2013 // indirect
	google.golang.org/grpc v1.44.0 // indirect
	google.golang.org/protobuf v1.25.0 // indirect
	// github.com/salrashid123/go-grpc-bazel-docker/echo v0.0.0
)

// replace github.com/salrashid123/go-grpc-bazel-docker/echo => ./echo
```


`C)`: Edit `echo/BUILD.bazel`

Enable the rule that uses `echo.pb.go` and disable the rest:

```bazel
# proto_library(
#     name = "echo_proto",
#     srcs = ["echo.proto"],
#     visibility = ["//visibility:public"],
# )

# go_proto_library(
#     name = "echo_go_proto",
#     compiler = "@io_bazel_rules_go//proto:go_grpc",
#     compilers = ["@io_bazel_rules_go//proto:go_grpc"],
#     importpath = "github.com/salrashid123/go-grpc-bazel-docker/echo",
#     proto = ":echo_proto",
#     visibility = ["//visibility:public"],
# )

# go_library(
#     name = "go_default_library",
#     embed = [":echo_go_proto"],
#     importpath = "echo",
#     visibility = ["//visibility:public"],
# )

go_library(
    name = "echo_go_proto",
    srcs = [
        "echo.pb.go",
    ],
    importpath = "github.com/salrashid123/go-grpc-bazel-docker/echo",
    visibility = ["//visibility:public"],
    deps = [
        "@com_github_golang_protobuf//proto:go_default_library",
        "@org_golang_google_protobuf//reflect/protoreflect:go_default_library",
        "@org_golang_google_protobuf//runtime/protoimpl:go_default_library",
        "@org_golang_google_grpc//:go_default_library",
        "@org_golang_google_grpc//codes:go_default_library",
        "@org_golang_google_grpc//status:go_default_library",
        "@org_golang_google_grpc//credentials:go_default_library",                          
        "@org_golang_x_net//context:go_default_library",         
    ],
)
```

`D)`  Run `gazelle` to populate dependencies in `repositories.bzl`:

```
bazel run :gazelle -- update-repos -from_file=go.mod -prune=true -to_macro=repositories.bzl%go_repositories
```

### Build and run without bazel

The focus of this repo is to use bazel to build and run. ...but if you want to manually build the proto and use go,

```bash
$ protoc --version
  libprotoc 3.19.1

$ go version
   go version go1.17.1 linux/amd64

$ bazel version
  Build label: 5.0.0

# as go, optionally compile
/usr/local/bin/protoc -I ./echo  \
  --include_imports --include_source_info \
  --descriptor_set_out=echo/echo.proto.pb \
  --go_opt=paths=source_relative \
  --go_out=plugins=grpc:./echo/ echo/echo.proto
```

Edit `go.mod` and uncomment the local imports.  The file should look like

```yaml
module main

go 1.17

require (
	github.com/golang/protobuf v1.4.3 // indirect
	github.com/google/uuid v1.3.0 // indirect
	golang.org/x/net v0.0.0-20220127200216-cd36cc0744dd // indirect
	golang.org/x/sys v0.0.0-20211216021012-1d35b9e2eb4e // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/genproto v0.0.0-20200526211855-cb27e3aa2013 // indirect
	google.golang.org/grpc v1.44.0 // indirect
	google.golang.org/protobuf v1.25.0 // indirect
	github.com/salrashid123/go-grpc-bazel-docker/echo v0.0.0
)

replace github.com/salrashid123/go-grpc-bazel-docker/echo => ./echo
```

then,

```
go run greeter_server/main.go --grpcport :50051  
go run greeter_client/main.go --host localhost:50051
```
