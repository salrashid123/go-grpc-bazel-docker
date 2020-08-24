# go with bazel with grpc with docker

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
bazel build  --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 examples/greeter_server:all
bazel build  --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 examples/greeter_client:all

bazel run  --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 examples/greeter_client:greeter_client_image
bazel run  --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 examples/greeter_server:greeter_server_image
```

### Check Image

```bash
$ docker images 
REPOSITORY                                       TAG                    IMAGE ID            CREATED             SIZE
bazel/examples/greeter_client                    greeter_client_image   a634c1e4b528        50 years ago        15.8MB
bazel/examples/greeter_server                    greeter_server_image   409954cce6b1        50 years ago        16MB
```

Notice the image hash values,...these wil be the same no matter where you generate the images

```yaml
$ docker inspect bazel/examples/greeter_client:greeter_client_image
[
    {
        "Id": "sha256:a634c1e4b528d66a57dc3ee2c30ca80750b66e7632f31958fd1f6b27937083b4",
        "RepoTags": [
            "bazel/examples/greeter_client:greeter_client_image"
        ],
        "RepoDigests": [],
        "Parent": "",
        "Comment": "",
        "Created": "1970-01-01T00:00:00Z",
        "Container": "e61c47f59d8323f7c6db62b1c47bb70faf0d8604756b85e9eb0cd329e88872d8",
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
            "Image": "sha256:197f4c09efde086f81e43b032e2677e02fa70ef01296ed28cdc69d1c5f8405b5",
            "Volumes": null,
            "WorkingDir": "",
            "Entrypoint": [
                "/client"
            ],
            "OnBuild": null,
            "Labels": null
        },
        "Architecture": "amd64",
        "Os": "linux",
        "Size": 15759905,
        "VirtualSize": 15759905,
        "GraphDriver": {
            "Data": {
                "LowerDir": "/var/lib/docker/overlay2/066c5b6ddf94a3b31f953b94453f43d5884afa06e2d8bd15a9d073685556c9e4/diff",
                "MergedDir": "/var/lib/docker/overlay2/64c753186f7206dd28f2eeaa8d32bf66230effa7901f1591d894588d12e51022/merged",
                "UpperDir": "/var/lib/docker/overlay2/64c753186f7206dd28f2eeaa8d32bf66230effa7901f1591d894588d12e51022/diff",
                "WorkDir": "/var/lib/docker/overlay2/64c753186f7206dd28f2eeaa8d32bf66230effa7901f1591d894588d12e51022/work"
            },
            "Name": "overlay2"
        },
        "RootFS": {
            "Type": "layers",
            "Layers": [
                "sha256:7444ea29e45e927abea1f923bf24cac20deaddea603c4bb1c7f2f5819773d453",
                "sha256:db047417f3aabafd9ff4e347aea4c113adf6be679601510c294ca0a3141b74db"
            ]
        },
        "Metadata": {
            "LastTagTime": "2020-08-24T15:15:48.872078308-04:00"
        }
    }
]
```

### (optional) gRPC Client/Server
```
docker run -p 50051:50051 bazel/examples/greeter_server:greeter_server_image
docker run --net=host bazel/examples/greeter_client:greeter_client_image
```

### TODO:

- use `gazelle` for dependencies `bazel run //:gazelle -- update-repos -from_file=examples/greeter_server/go.mod`
- Setup Bazel with [Google Cloud Build](https://cloud.google.com/cloud-build/docs/cloud-builders)
