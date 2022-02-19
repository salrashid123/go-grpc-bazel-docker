# Deterministic builds with go + bazel + grpc + docker

Deterministic container images for gRPC+golang [bazel](https://bazel.build/).

The following sample will build a golang gRPC client/server and then embed the binaries into container images.

These images are will have a consistent image hash no matter where it is built

*  `greeter_server@sha256:b1efcd6fcc04cf812f49b42dd6430b2a745f486e85d4f53cd69379509c9de8b3`
*  `greeter_client@sha256:72cf416671f3722cd69c6b48a6f6faf134ec1679904b0371c491315f42f54913`

For reference, see:

- [Building deterministic Docker images with Bazel](https://blog.bazel.build/2015/07/28/docker_build.html)
- [Create Container images with Bazel](https://dev.to/schoren/create-container-images-with-bazel-47am)
- [rules_docker](https://github.com/bazelbuild/rules_docker)


To run this sample, you will need `bazel` installed (see [Cloud Shell](#cloud-shell) for an easy way to use `bazel`)

In the end, you'll end up with the same digests

```bash
$ docker pull salrashid123/greeter_server:greeter_server_image
$ docker inspect salrashid123/greeter_server:greeter_server_image

[
    {
        "RepoTags": [
            "bazel/greeter_server:greeter_server_image",
            "salrashid123/greeter_server:greeter_server_image"
        ],
        "RepoDigests": [
            "salrashid123/greeter_server@sha256:b1efcd6fcc04cf812f49b42dd6430b2a745f486e85d4f53cd69379509c9de8b3"
        ],

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

Note, the bazel version used here is 

```bash
$ docker run gcr.io/cloud-builders/bazel@sha256:4a19236baf0e5d663942c3947497e3f5b5356ae3dd6f97b1fae92897a97a11ad version
    Build label: 5.0.0
    Build target: bazel-out/k8-opt/bin/src/main/java/com/google/devtools/build/lib/bazel/BazelServer_deploy.jar
    Build time: Wed Jan 19 14:08:54 2022 (1642601334)
    Build timestamp: 1642601334
    Build timestamp as int: 1642601334
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
bazel run :gazelle -- update-repos -from_file=go.mod -prune=true -to_macro=repositories.bzl%go_repositories
```

Then build the client and server

```bash
bazel build --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 greeter_server:all
bazel run  --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 greeter_server:greeter_server_image

bazel build --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 greeter_client:all
bazel run  --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 greeter_client:greeter_client_image
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

#### without TLS

- with docker

```
docker run -p 50051:50051 bazel/greeter_server:greeter_server_image --grpcport :50051 --insecure
docker run --network="host" bazel/greeter_client:greeter_client_image --host localhost:50051 --insecure -skipHealthCheck 
```

with bazel

```
bazel run --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 greeter_server:server -- --grpcport :50051 --insecure
bazel run --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 greeter_client:client -- --host localhost:50051 --insecure -skipHealthCheck
```

#### with TLS

- with docker

```bash
docker run -v `pwd`/certs:/certs \
  -p 50051:50051 bazel/greeter_server:greeter_server_image  \
  -grpcport :50051 \
  --tlsCert certs/grpc_server_crt.pem \
  --tlsKey certs/grpc_server_key.pem

docker run -v `pwd`/certs:/certs \
  --network="host" bazel/greeter_client:greeter_client_image \
  --host localhost:50051 \
  --tlsCert certs/CA_crt.pem \
  --servername grpc.domain.com \
  -skipHealthCheck
```
with bazel

```bash
bazel run --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 greeter_server:server \
  -- --grpcport :50051 \
  --tlsCert certs/grpc_server_crt.pem \
  --tlsKey certs/grpc_server_key.pem

bazel run --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 greeter_client:client \
  -- \
  --host localhost:50051 \
  --tlsCert certs/CA_crt.pem \
  --servername grpc.domain.com \
  -skipHealthCheck
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
go run greeter_server/main.go --grpcport :50051 \
  --tlsCert certs/grpc_server_crt.pem \
  --tlsKey certs/grpc_server_key.pem

go run greeter_client/main.go \
  --host localhost:50051 \
  --tlsCert certs/CA_crt.pem \
  --servername grpc.domain.com \
  -skipHealthCheck
```

### Specify docker image

Specify a docker repo to by setting the `repository` command here. In the case below, its container registry `gcr.io/project_id`

```bazel
container_image(
    name = "greeter_server_image",
    base = "@alpine_linux_amd64//image",
    entrypoint = ["/server"],
    files = [":server"],
    repository = "gcr.io/project_id`"
)
```

on push to a repo

- `Server`
```bash
$ bazel build --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 greeter_server:all
$ bazel run  --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 greeter_server:greeter_server_image

$ docker push gcr.io/project_id/greeter_server:greeter_server_image
    greeter_server_image: digest: sha256:b1efcd6fcc04cf812f49b42dd6430b2a745f486e85d4f53cd69379509c9de8b3 size: 738
```

you'll see the hash we need...this is specific and intrinsic to the image.

On any other machine, generate the builds and inspect

```bash
$ docker pull salrashid123/greeter_server:greeter_server_image
$ docker inspect gcr.io/project_id/greeter_server:greeter_server_image
```


### Cloud Build

You can use Cloud Build to create the image by using the `bazel` builder and specifying the repository path to export to.  In the sample below, the repository is set o google container registry:

```yaml
container_image(
    name = "greeter_server_image",
    base = "@alpine_linux_amd64//image",
    entrypoint = ["/server"],
    files = [":server"],
    repository = "gcr.io/your_project"
)
```

Note that `cloudbuild.yaml` specifies the base bazel version by hash too

```yaml
steps:
- name: gcr.io/cloud-builders/bazel@sha256:4a19236baf0e5d663942c3947497e3f5b5356ae3dd6f97b1fae92897a97a11ad
  args: ['run', '--platforms=@io_bazel_rules_go//go/toolchain:linux_amd64', 'greeter_server:greeter_server_image']
images: ['gcr.io/$PROJECT_ID/greeter_server:greeter_server_image']
```


```bash
$ bazel clean
$ gcloud builds submit --config=cloudbuild.yaml --machine-type=n1-highcpu-32

    Tagging 67ccf97f942115c222d219c3d188570972e966feaccc01a6c117a3a88b14de9d as gcr.io/mineral-minutia-820/greeter_server:greeter_server_image
    PUSH
    Pushing gcr.io/mineral-minutia-820/greeter_server:greeter_server_image
    The push refers to repository [gcr.io/mineral-minutia-820/greeter_server]
    2b2aace28f76: Preparing
    5d09c2db1d76: Preparing
    417cb9b79ade: Preparing
    417cb9b79ade: Pushed
    2b2aace28f76: Pushed
    5d09c2db1d76: Pushed
    greeter_server_image: digest: sha256:b1efcd6fcc04cf812f49b42dd6430b2a745f486e85d4f53cd69379509c9de8b3 size: 948
    DONE
    ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
    ID                                    CREATE_TIME                DURATION  SOURCE                                                                                            IMAGES                                                          STATUS
    cba32e2f-93ac-4a82-aaba-c6d231259a97  2022-02-19T19:56:09+00:00  2M4S      gs://mineral-minutia-820_cloudbuild/source/1645300568.75279-f059d603e657465ea770201440b1f935.tgz  gcr.io/mineral-minutia-820/greeter_server:greeter_server_image  SUCCESS

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
go run greeter_server/main.go --grpcport :50051 --insecure  
go run greeter_client/main.go --host localhost:50051 --insecure
```
