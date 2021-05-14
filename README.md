# Deterministic builds with go + bazel + grpc + docker

Deterministic container images for gRPC+golang [bazel](https://bazel.build/).

The following sample will build a golang gRPC client/server and then embed the binaries into container images.

These images are will have a consistent image hash no matter where it is built (eg, `greeter_server@sha256:2cb462befb6eed81508a098452ebfa920e2305cde48e1ecc3a56efde75912360`)

For reference, see:

- [Building deterministic Docker images with Bazel](https://blog.bazel.build/2015/07/28/docker_build.html)
- [Create Container images with Bazel](https://dev.to/schoren/create-container-images-with-bazel-47am)
- [rules_docker](https://github.com/bazelbuild/rules_docker)


To run this sample, you will need `bazel` installed (see [Cloud Shell](#cloud-shell) for an easy way to use `bazel`)

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
    Build label: 4.0.0
    Build target: bazel-out/k8-opt/bin/src/main/java/com/google/devtools/build/lib/bazel/BazelServer_deploy.jar
    Build time: Thu Jan 21 07:33:24 2021 (1611214404)
    Build timestamp: 1611214404
    Build timestamp as int: 1611214404

# or 

$ docker run gcr.io/cloud-builders/bazel@sha256:0bb18b771de34c386ae26bfac960cd57fda889eeef1f0171e10dab73e17cade3 version
    Build label: 4.0.0
    Build target: bazel-out/k8-opt/bin/src/main/java/com/google/devtools/build/lib/bazel/BazelServer_deploy.jar
    Build time: Thu Jan 21 07:33:24 2021 (1611214404)
    Build timestamp: 1611214404
    Build timestamp as int: 1611214404
```

### Build Image

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
    bazel/greeter_server          greeter_server_image   a5315b9825fb   51 years ago   17.3MB
    bazel/greeter_client          greeter_client_image   014df8e803e8   51 years ago   17.1MB
```

Inspect the image thats generated.  The hash we're after is actually `RepoTags` which we'll generate and show later, for now

### (optional) Run the gRPC Client/Server

(why not, you already built it)

#### no TLS

with go

```bash
# as go, optionally compile
# /usr/local/bin/protoc -I ./echo \
#   --include_imports --include_source_info \
#   --descriptor_set_out=echo/echo.proto.pb \
#   --go_out=plugins=grpc:./echo/ echo/echo.proto

go run greeter_server/main.go --grpcport :50051 --insecure  
go run greeter_client/main.go --host localhost:50051 --insecure
```

with docker:

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

with go

```bash
go run greeter_server/main.go --grpcport :50051 --tlsCert certs/grpc_server_crt.pem --tlsKey certs/grpc_server_key.pem
go run greeter_client/main.go --host localhost:50051 --tlsCert certs/CA_crt.pem --servername grpc.domain.com -skipHealthCheck
```

with docker

```bash
docker run -v `pwd`/certs:/certs -p 50051:50051 bazel/greeter_server:greeter_server_image --grpcport :50051 --tlsCert certs/grpc_server_crt.pem --tlsKey certs/grpc_server_key.pem
docker run -v `pwd`/certs:/certs --network="host" bazel/greeter_client:greeter_client_image --host localhost:50051 --tlsCert certs/CA_crt.pem --servername grpc.domain.com -skipHealthCheck
```
with bazel

```bash
bazel run --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 greeter_server:server -- --grpcport :50051 --tlsCert certs/grpc_server_crt.pem --tlsKey certs/grpc_server_key.pem
bazel run --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 greeter_client:client -- --host localhost:50051 --tlsCert certs/CA_crt.pem --servername grpc.domain.com -skipHealthCheck
```

### Specify docker image

Specify a docker repo to by setting the `repository` command here. In the case below, its on dockerhub as handle `salrashid123`

```bazel
container_image(
    name = "greeter_server_image",
    base = "@alpine_linux_amd64//image",
    entrypoint = ["/server"],
    files = [":server"],
    repository = "salrashid123"
)
```

on push to a repo

- `Server`
```bash
$ bazel build --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 greeter_server:all
$ bazel run  --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 greeter_server:greeter_server_image

$ docker push salrashid123/greeter_server:greeter_server_image
    greeter_server_image: digest: sha256:2cb462befb6eed81508a098452ebfa920e2305cde48e1ecc3a56efde75912360 size: 738

```

you'll see the hash we need...this is specific and intrinsic to the image.

On any other machine pull the image and inspect

```bash
$ docker inspect salrashid123/greeter_server@sha256:2cb462befb6eed81508a098452ebfa920e2305cde48e1ecc3a56efde75912360

[
    {
        "Id": "sha256:a5315b9825fbaa72d36e4dd7a665e51120821dd89149a439dfde4f2e271889e4",
        "RepoTags": [
            "bazel/greeter_server:greeter_server_image",
            "gcr.io/mineral-minutia-820/greeter_server:greeter_server_image"
        ],
        "RepoDigests": [
            "salrashid123/greeter_server@sha256:2cb462befb6eed81508a098452ebfa920e2305cde48e1ecc3a56efde75912360",
            "gcr.io/mineral-minutia-820/greeter_server@sha256:2cb462befb6eed81508a098452ebfa920e2305cde48e1ecc3a56efde75912360"
        ],
   ...
```


### Cloud Build

You can use Cloud Build to create the image by using the `bazel` builder and specifying the repository path to export to.  In the sample below, the repository is set o google container registry:

```
container_image(
    name = "greeter_server_image",
    base = "@alpine_linux_amd64//image",
    entrypoint = ["/server"],
    files = [":server"],
    repository = "gcr.io/mineral-minutia-820"
)
```

```bash
$ bazel clean
$ gcloud builds submit --config=cloudbuild.yaml --machine-type=n1-highcpu-32



```

Note the docker hub image hash and gcr.io hash for the server is

`sha256:2cb462befb6eed81508a098452ebfa920e2305cde48e1ecc3a56efde75912360`





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
$ docker run gcr.io/cloud-builders/bazel@sha256:77d42f9b252c6b159416b5651dd69a7861d0f2ffbea05bc5b6482caf846ec9f4 version

```

### Using Pregenerated protopb and gazelle

`A)` Generate `proto.pb`:

```
/usr/local/bin/protoc -I ./echo \
  --include_imports --include_source_info \
  --descriptor_set_out=echo/echo.proto.pb \
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
//	echo v0.0.0
)

//replace echo => ./echo

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
#     importpath = "echo",
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
    importpath = "echo",
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

`D)`  Run `gazelle` to populate dependencies in `WORKSPACE`:

```
bazel run :gazelle -- update-repos -from_file=go.mod -build_file_proto_mode=disable_global
```