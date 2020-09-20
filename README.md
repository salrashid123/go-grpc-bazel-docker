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

Note, the bazel base image specifies the image+hash so your'e starting off from a known state:

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
$ docker images  | grep bazel
bazel/greeter_server                        greeter_server_image   2de507b33df9        50 years ago        16MB
bazel/greeter_client                        greeter_client_image   0ebd575f57c5        50 years ago        15.8MB
```

Inspect the image thats generated...these wil be the same no matter where you generate the images

```yaml
$ docker inspect bazel/greeter_server:greeter_server_image
[
    {
        "Id": "sha256:2de507b33df9e0ffee6cb883c318a4d4fc0b487e6fae04ee1be1120f8f5a329c",
        "RepoTags": [
            "bazel/greeter_server:greeter_server_image"
        ],
        "RepoDigests": [],
        "Parent": "",
        "Comment": "",
        "Created": "1970-01-01T00:00:00Z",
        "Container": "f382632c7b88c2348c28b8754b3aeb69f3c69448d48a7c8e27675abd309a04cf",
        "ContainerConfig": {
            "Hostname": "",
            "Domainname": "",
            "User": "",
            "AttachStdin": false,
            "AttachStdout": false,
            "AttachStderr": false,
            "Tty": false,
            "OpenStdin": false,
            "StdinOnce": false,
            "Env": null,
            "Cmd": null,
            "Image": "",
            "Volumes": null,
            "WorkingDir": "",
            "Entrypoint": null,
            "OnBuild": null,
            "Labels": null
        },
        "DockerVersion": "18.06.1-ce",
        "Author": "Bazel",
        "Config": {
            "Hostname": "",
            "Domainname": "",
            "User": "",
            "AttachStdin": false,
            "AttachStdout": false,
            "AttachStderr": false,
            "Tty": false,
            "OpenStdin": false,
            "StdinOnce": false,
            "Env": [
                "PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
            ],
            "Cmd": [
                "/bin/sh"
            ],
            "ArgsEscaped": true,
            "Image": "sha256:57e798527dbcbda6abbc2214fe12346873f25ba65c6ced0a65a149b316a3e9a1",
            "Volumes": null,
            "WorkingDir": "",
            "Entrypoint": [
                "/server"
            ],
            "OnBuild": null,
            "Labels": null
        },
        "Architecture": "amd64",
        "Os": "linux",
        "Size": 16022581,
        "VirtualSize": 16022581,
        "GraphDriver": {
            "Data": {
                "LowerDir": "/var/lib/docker/overlay2/567091aeb276db80ba6894ced4fe82b4761108a9e0433b1cca470c85686bd194/diff",
                "MergedDir": "/var/lib/docker/overlay2/8d6a205645512e04ff2f5002057c53a60ef99537834e7d85de1fcd597e2fc3bb/merged",
                "UpperDir": "/var/lib/docker/overlay2/8d6a205645512e04ff2f5002057c53a60ef99537834e7d85de1fcd597e2fc3bb/diff",
                "WorkDir": "/var/lib/docker/overlay2/8d6a205645512e04ff2f5002057c53a60ef99537834e7d85de1fcd597e2fc3bb/work"
            },
            "Name": "overlay2"
        },
        "RootFS": {
            "Type": "layers",
            "Layers": [
                "sha256:a1852e9ff2e7cb61bd911cb964ae939e95621121a53b1e5af7c2cb341cd04283",
                "sha256:3a3b796daf68cc68a197c363e60b22adbb10e383a22bdbeb165f96aa02cb1b4c"
            ]
        },
        "Metadata": {
            "LastTagTime": "2020-09-20T12:09:22.440811453-04:00"
        }
    }
]
```

### (optional) gRPC Client/Server

(why not?)
```
docker run -p 50051:50051 bazel/greeter_server:greeter_server_image
docker run --network="host" bazel/greeter_client:greeter_client_image
```

or directly with bazel

```
bazel run --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 greeter_server:server
bazel run --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 greeter_client:client
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
    greeter_server_image: digest: sha256:6e10797105b2bd4889ac10155dcb87dc615d4168fbaf78444e8ec6fdbaf7a967 size: 738

```

On any other machine pull the image and inspect

```bash
$ docker inspect salrashid123/greeter_server@sha256:6e10797105b2bd4889ac10155dcb87dc615d4168fbaf78444e8ec6fdbaf7a967
[
    {
        "Id": "sha256:2de507b33df9e0ffee6cb883c318a4d4fc0b487e6fae04ee1be1120f8f5a329c",
        "RepoTags": [
            "bazel/greeter_server:greeter_server_image",
            "salrashid123/greeter_server:greeter_server_image"
        ],
        "RepoDigests": [
            "salrashid123/greeter_server@sha256:6e10797105b2bd4889ac10155dcb87dc615d4168fbaf78444e8ec6fdbaf7a967"
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

    Loaded image ID: sha256:2de507b33df9e0ffee6cb883c318a4d4fc0b487e6fae04ee1be1120f8f5a329c
    Tagging 2de507b33df9e0ffee6cb883c318a4d4fc0b487e6fae04ee1be1120f8f5a329c as gcr.io/mineral-minutia-820/greeter_server:greeter_server_image
    PUSH
    greeter_server_image: digest: sha256:6e10797105b2bd4889ac10155dcb87dc615d4168fbaf78444e8ec6fdbaf7a967 size: 738
    DONE

    ID                                    CREATE_TIME                DURATION  SOURCE                                                                                             IMAGES                                                          STATUS
    d9b44831-e2b5-459b-b5d1-83c78a30a707  2020-09-20T16:14:34+00:00  2M25S     gs://mineral-minutia-820_cloudbuild/source/1600618473.457799-5b397e0230844adbb7b2f13721b24cf4.tgz  gcr.io/mineral-minutia-820/greeter_server:greeter_server_image  SUCCESS
```

Note the docker hub image hash and gcr.io hash for the server is

`sha256:6e10797105b2bd4889ac10155dcb87dc615d4168fbaf78444e8ec6fdbaf7a967`

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