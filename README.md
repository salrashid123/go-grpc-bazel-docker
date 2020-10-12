# Deterministic builds with go + bazel + grpc + docker

Deterministic container images for gRPC+golang [bazel](https://bazel.build/).

The following sample will build a golang gRPC client/server and then embed the binaries into container images.

These images are will have a consistent image hash no matter where it is built.

For reference, see:

- [Building deterministic Docker images with Bazel](https://blog.bazel.build/2015/07/28/docker_build.html)
- [Create Container images with Bazel](https://dev.to/schoren/create-container-images-with-bazel-47am)
- [rules_docker](https://github.com/bazelbuild/rules_docker)


To run this sample, you will need `bazel` installed


### Build Image

```bash
bazel build --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 greeter_server:all
bazel run  --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 greeter_server:greeter_server_image

bazel build --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 greeter_client:all
bazel run  --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 greeter_client:greeter_client_image
```

Note, the bazel base image specifies the image+hash so you're starting off from a known state:

- `WORKSPACE`

```
container_pull(
    name = "alpine_linux_amd64",
    registry = "index.docker.io",
    repository = "library/alpine",
    tag = "3.8",
    digest = "sha256:cf35b4fa14e23492df67af08ced54a15e68ad00cac545b437b1994340f20648c"
)
```

### Check Image

```bash
bazel/greeter_client                         greeter_client_image                     c44e11355e04        50 years ago        15.9MB
bazel/greeter_server                         greeter_server_image                     9dcc3f1692fe        50 years ago        16.1MB
```

Inspect the image thats generated...these wil be the same no matter where you generate the images

```yaml
$ docker inspect bazel/greeter_server:greeter_server_image
```

### (optional) gRPC Client/Server

(why not?)
```
docker run -p 50051:50051 bazel/greeter_server:greeter_server_image --grpcport :50051
docker run --network="host" bazel/greeter_client:greeter_client_image --host localhost:50051
```

or directly with bazel

```
bazel run --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 greeter_server:server -- --grpcport :50051
bazel run --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 greeter_client:client -- --host localhost:50051
```

### Specify docker image

You can specify a repo prefix by setting the `repository` command here. In the case below, its on dockerhub as handle `salrashid123`

```bazel
container_image(
    name = "greeter_server_image",
    base = "@alpine_linux_amd64//image",
    entrypoint = ["/server"],
    files = [":server"],
    repository = "salrashid123"
)
```

on push to dockerhub

- `Server`
```bash
$ bazel build --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 greeter_server:all
$ bazel run  --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 greeter_server:greeter_server_image

$ docker push salrashid123/greeter_server:greeter_server_image
    greeter_server_image: digest: sha256:99be982551073de4dd1989940c80bdedcd84b1a0b42eb76e953a3f1b1643e56b size: 738

```

On any other machine pull the image and inspect

```bash
$ docker inspect salrashid123/greeter_server@sha256:99be982551073de4dd1989940c80bdedcd84b1a0b42eb76e953a3f1b1643e56b

        "Id": "sha256:9dcc3f1692fe1eee9cfdd934c64e0af025cda2ddd13b442ec712b96f3b5576ca",
        "RepoTags": [
            "bazel/greeter_server:greeter_server_image",
            "salrashid123/greeter_server:greeter_server_image"
        ],
        "RepoDigests": [
            "salrashid123/greeter_server@sha256:99be982551073de4dd1989940c80bdedcd84b1a0b42eb76e953a3f1b1643e56b"
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
$ gcloud builds submit --config=cloudbuild.yaml --machine-type=n1-highcpu-32

        Loaded image ID: sha256:9dcc3f1692fe1eee9cfdd934c64e0af025cda2ddd13b442ec712b96f3b5576ca
        Tagging 9dcc3f1692fe1eee9cfdd934c64e0af025cda2ddd13b442ec712b96f3b5576ca as gcr.io/mineral-minutia-820/greeter_server:greeter_server_image
        PUSH
        Pushing gcr.io/mineral-minutia-820/greeter_server:greeter_server_image
        The push refers to repository [gcr.io/mineral-minutia-820/greeter_server]
        greeter_server_image: digest: sha256:99be982551073de4dd1989940c80bdedcd84b1a0b42eb76e953a3f1b1643e56b size: 738
        DONE
        -----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

        ID                                    CREATE_TIME                DURATION  SOURCE                                                                                             IMAGES                                                          STATUS
        ba9d078e-d974-458b-afe8-af1fdf271638  2020-10-12T15:03:39+00:00  2M13S     gs://mineral-minutia-820_cloudbuild/source/1602515018.756311-f40d9688cca14064afb512e4b0fa576e.tgz  gcr.io/mineral-minutia-820/greeter_server:greeter_server_image  SUCCESS
```

Note the docker hub image hash and gcr.io hash for the server is

`sha256:ad143241dbe86f462d73006acb8c70119da319f2b2a8c6da6881d7a6e6e21a9b`

### Using Pregenerated protopb and gazelle

`A)` Generate `proto.pb`:

```
/usr/local/bin/protoc -I ./helloworld \
  --include_imports --include_source_info \
  --descriptor_set_out=helloworld/helloworld.proto.pb \
  --go_out=plugins=grpc:./helloworld/ helloworld/helloworld.proto
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
//	helloworld v0.0.0
)

//replace helloworld => ./helloworld

```


`C)`: Edit `helloworld/BUILD.bazel`

Enable the rule that uses `helloworld.pb.go` and disable the rest:

```bazel
proto_library(
#     name = "helloworld_proto",
#     srcs = ["helloworld.proto"],
#     visibility = ["//visibility:public"],
# )

# go_proto_library(
#     name = "helloworld_go_proto",
#     compiler = "@io_bazel_rules_go//proto:go_grpc",
#     compilers = ["@io_bazel_rules_go//proto:go_grpc"],
#     importpath = "helloworld",
#     proto = ":helloworld_proto",
#     visibility = ["//visibility:public"],
# )

# go_library(
#     name = "go_default_library",
#     embed = [":helloworld_go_proto"],
#     importpath = "helloworld",
#     visibility = ["//visibility:public"],
# )

go_library(
    name = "helloworld_go_proto",
    srcs = [
        "helloworld.pb.go",
    ],
    importpath = "helloworld",
    visibility = ["//visibility:public"],
    deps = [
        "@com_github_golang_protobuf//proto:go_default_library",
        "@org_golang_google_protobuf//reflect/protoreflect:go_default_library",
        "@org_golang_google_protobuf//runtime/protoimpl:go_default_library",
        "@org_golang_google_grpc//:go_default_library",
        "@org_golang_google_grpc//codes:go_default_library",
        "@org_golang_google_grpc//status:go_default_library",                
        "@org_golang_x_net//context:go_default_library",         
    ],
)
```

`D)`  Run `gazelle` to populate dependencies in `WORKSPACE`:

```
bazel run :gazelle -- update-repos -from_file=go.mod -build_file_proto_mode=disable_global
```