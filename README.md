# Deterministic builds with go + bazel + grpc + docker

Deterministic container images for gRPC+golang [bazel](https://bazel.build/).

The following sample will build a golang gRPC client/server and then embed the binaries into container images.

These images are will have a consistent image hash no matter where it is built

*  `greeter_server@sha256:6a296dbc5b96b051554ec589648fbdd83993cf5889630cd4098606c8fd7d0eac`
*  `greeter_client@sha256:92251eecfb9dfbbaba55b772353d970bf3b446ca598b3f543704950e7a75f47a`

For reference, see:

- [Building deterministic Docker images with Bazel](https://blog.bazel.build/2015/07/28/docker_build.html)
- [Create Container images with Bazel](https://dev.to/schoren/create-container-images-with-bazel-47am)
- [rules_docker](https://github.com/bazelbuild/rules_docker)


To run this sample, you will need `bazel` installed (see [Cloud Shell](#cloud-shell) for an easy way to use `bazel`)

In the end, you'll end up with the same digests

```bash
$ docker pull salrashid123/greeter_server:greeter_server_image
$ docker pull gcr.io/your_project/greeter_server:greeter_server_image

$ docker inspect gcr.io/your_project/greeter_server:greeter_server_image
[
    {
        "Id": "sha256:8668f16460afb698fbfe2d15f4efe13cda0d56b73a5d78ca26e9e5a23e74de29",
        "RepoTags": [
            "bazel/greeter_server:greeter_server_image",
            "salrashid123/greeter_server:greeter_server_image",
            "gcr.io/mineral-minutia-820/greeter_server:greeter_server_image"
        ],
        "RepoDigests": [
            "salrashid123/greeter_server@sha256:6a296dbc5b96b051554ec589648fbdd83993cf5889630cd4098606c8fd7d0eac",
            "gcr.io/mineral-minutia-820/greeter_server@sha256:6a296dbc5b96b051554ec589648fbdd83993cf5889630cd4098606c8fd7d0eac"
        ],

   ...
```

### Cloud Shell

If you have access to Google Cloud Platform account, you can use Cloud Shell to run `bazel` and save yourself an installation.

```bash
gcloud alpha cloud-shell ssh 

git clone https://github.com/salrashid123/go-grpc-bazel-docker.git
cd go-grpc-bazel-docker
```

Then within the shell, you should be able to `bazel version` to ensure it is installed.


```bash
$ bazel version
    Build label: 4.1.0
    Build target: bazel-out/k8-opt/bin/src/main/java/com/google/devtools/build/lib/bazel/BazelServer_deploy.jar
    Build time: Fri May 21 11:11:34 2021 (1621595494)
    Build timestamp: 1621595494
    Build timestamp as int: 1621595494

# or 

$ docker run gcr.io/cloud-builders/bazel@sha256:0bb18b771de34c386ae26bfac960cd57fda889eeef1f0171e10dab73e17cade3 version
    Build label: 4.1.0
    Build target: bazel-out/k8-opt/bin/src/main/java/com/google/devtools/build/lib/bazel/BazelServer_deploy.jar
    Build time: Fri May 21 11:11:34 2021 (1621595494)
    Build timestamp: 1621595494
    Build timestamp as int: 1621595494
```

### Build Image with Bazel

Declare go dependencies from `go.mod` into `repositories.bzl`:

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
REPOSITORY                    TAG                    IMAGE ID       CREATED        SIZE
bazel/greeter_client          greeter_client_image   aad682ca02bf   51 years ago    17.1MB
bazel/greeter_server          greeter_server_image   cfc5df5beab8   51 years ago    17.4MB
```

Inspect the image thats generated.  The hash we're after is actually `RepoTags` which we'll generate and show later, for now

### (optional) Run the gRPC Client/Server

(why not, you already built it)

#### withoug TLS

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
docker run -v `pwd`/certs:/certs -p 50051:50051 bazel/greeter_server:greeter_server_image --grpcport :50051 --tlsCert certs/grpc_server_crt.pem --tlsKey certs/grpc_server_key.pem

docker run -v `pwd`/certs:/certs --network="host" bazel/greeter_client:greeter_client_image --host localhost:50051 --tlsCert certs/CA_crt.pem --servername grpc.domain.com -skipHealthCheck
```
with bazel

```bash
bazel run --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 greeter_server:server -- --grpcport :50051 --tlsCert certs/grpc_server_crt.pem --tlsKey certs/grpc_server_key.pem

bazel run --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 greeter_client:client -- --host localhost:50051 --tlsCert certs/CA_crt.pem --servername grpc.domain.com -skipHealthCheck
```

with go

You will first want to build the files, see corresponding steps above

```bash
go run greeter_server/main.go --grpcport :50051 --tlsCert certs/grpc_server_crt.pem --tlsKey certs/grpc_server_key.pem
go run greeter_client/main.go --host localhost:50051 --tlsCert certs/CA_crt.pem --servername grpc.domain.com -skipHealthCheck
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
    greeter_server_image: digest: sha256:6a296dbc5b96b051554ec589648fbdd83993cf5889630cd4098606c8fd7d0eac size: 738
```

you'll see the hash we need...this is specific and intrinsic to the image.

On any other machine, generate the builds and inspect

```bash
$ docker pull salrashid123/greeter_server:greeter_server_image
$ docker inspect gcr.io/project_id/greeter_server:greeter_server_image

[
    {
        "Id": "sha256:8668f16460afb698fbfe2d15f4efe13cda0d56b73a5d78ca26e9e5a23e74de29",
        "RepoTags": [
            "bazel/greeter_server:greeter_server_image",
            "salrashid123/greeter_server:greeter_server_image",
            "gcr.io/project_id/greeter_server:greeter_server_image"
        ],
        "RepoDigests": [
            "salrashid123/greeter_server@sha256:6a296dbc5b96b051554ec589648fbdd83993cf5889630cd4098606c8fd7d0eac",
            "gcr.io/project_id/greeter_server@sha256:6a296dbc5b96b051554ec589648fbdd83993cf5889630cd4098606c8fd7d0eac"
        ],
   ...
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
- name: gcr.io/cloud-builders/bazel@sha256:36ab6b8816e473592fa70e0dd866caf0267cacc1ed6ac40266a082f0b70270a0
  args: ['run', '--platforms=@io_bazel_rules_go//go/toolchain:linux_amd64', 'greeter_server:greeter_server_image']
images: ['gcr.io/$PROJECT_ID/greeter_server:greeter_server_image']
```


```bash
$ bazel clean
$ gcloud builds submit --config=cloudbuild.yaml --machine-type=n1-highcpu-32

    ...
    ...
    Pushing gcr.io/project_id/greeter_server:greeter_server_image
    The push refers to repository [gcr.io/your_project/greeter_server]

    greeter_server_image: digest: sha256:6a296dbc5b96b051554ec589648fbdd83993cf5889630cd4098606c8fd7d0eac size: 738
    DONE

    ID                                    CREATE_TIME                DURATION  SOURCE                                                                                             IMAGES                                                          STATUS
    b4148461-6a57-485c-ae0f-08faa64b2e05  2021-06-25T13:47:29+00:00  2M14S     gs://project_id_cloudbuild/source/1624628848.805103-fb510921a1d848159fe2bf618c87f761.tgz  gcr.io/project_id/greeter_server:greeter_server_image  SUCCESS

```


### Attesting base dependencies

The `WORKSPACE` and git dependencies are all known down to the specific version of bazel and base container image

- `WORKSPACE`

The base image used for the client and server is alpine:

```
container_pull(
    name = "alpine_linux_amd64",
    registry = "index.docker.io",
    repository = "library/alpine",
    tag = "3.8",
    digest = "sha256:cf35b4fa14e23492df67af08ced54a15e68ad00cac545b437b1994340f20648c"
)
```

- Bazel note the version of bazel we are using:

```bash
$ docker run gcr.io/cloud-builders/bazel@sha256:36ab6b8816e473592fa70e0dd866caf0267cacc1ed6ac40266a082f0b70270a0 version
    Build label: 4.1.0
    Build target: bazel-out/k8-opt/bin/src/main/java/com/google/devtools/build/lib/bazel/BazelServer_deploy.jar
    Build time: Fri May 21 11:11:34 2021 (1621595494)
    Build timestamp: 1621595494
    Build timestamp as int: 1621595494
```

### Using Pregenerated protopb and gazelle

The default bazel configuration in `echo/BUILD.bazel` compiles the proto files.  If you would rather use pregenerated proto files (eg, to [avoid conflicts](https://github.com/bazelbuild/rules_go/blob/master/proto/core.rst#avoiding-conflicts), you must do that outside of bazel and just specify a library)

`A)` Generate `proto.pb`:

```
/usr/local/bin/protoc -I ./echo  \
  --include_imports --include_source_info \
  --descriptor_set_out=echo/echo.proto.pb \
  --go_opt=paths=source_relative \
  --go_out=plugins=grpc:./echo/ echo/echo.proto
```

`B)` comment the local `replace` directives in `go.mod`:

```
module main

go 1.14

require (
	github.com/google/uuid v1.1.2 // indirect
	golang.org/x/net v0.0.0-20200822124328-c89045814202 // indirect
	google.golang.org/grpc v1.31.1 // indirect
	google.golang.org/protobuf v1.25.0 // indirect
//	github.com/salrashid123/go-grpc-bazel-docker/echo v0.0.0
)

//replace github.com/salrashid123/go-grpc-bazel-docker/echo => ./echo

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
  libprotoc 3.13.0

$ go version
   go version go1.15.2 linux/amd64

$ bazel version
  Build label: 4.1.0

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

go 1.14

require (
	github.com/google/uuid v1.1.2 // indirect
	golang.org/x/net v0.0.0-20200822124328-c89045814202 // indirect
	google.golang.org/grpc v1.31.1 // indirect
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