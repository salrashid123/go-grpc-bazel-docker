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
bazel build  --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 greeter_server:all
bazel run  --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 greeter_server:greeter_server_image

bazel build  --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 greeter_client:all
bazel run  --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 greeter_client:greeter_client_image
```

### Check Image

```bash
$ docker images 
REPOSITORY                                       TAG                    IMAGE ID            CREATED             SIZE
bazel/greeter_client   greeter_client_image   0c4a07ae6d50        50 years ago        15.7MB
bazel/greeter_server   greeter_server_image   7ea0fc3e14c0        50 years ago        16MB
```

Inspect the image thats generated...these wil be the same no matter where you generate the images

```yaml
$ docker inspect bazel/greeter_server:greeter_server_image
[
    {
        "Id": "sha256:7ea0fc3e14c0d0cfdd8048f5ddd19566a1b78f822658b8c5318c14241340a982",
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
        "Size": 15971699,
        "VirtualSize": 15971699,
        "GraphDriver": {
            "Data": {
                "LowerDir": "/var/lib/docker/overlay2/567091aeb276db80ba6894ced4fe82b4761108a9e0433b1cca470c85686bd194/diff",
                "MergedDir": "/var/lib/docker/overlay2/c610376ebfd59f918f445ea74b2dde32b584a34fdf5a506b3845d2c47e069c52/merged",
                "UpperDir": "/var/lib/docker/overlay2/c610376ebfd59f918f445ea74b2dde32b584a34fdf5a506b3845d2c47e069c52/diff",
                "WorkDir": "/var/lib/docker/overlay2/c610376ebfd59f918f445ea74b2dde32b584a34fdf5a506b3845d2c47e069c52/work"
            },
            "Name": "overlay2"
        },
        "RootFS": {
            "Type": "layers",
            "Layers": [
                "sha256:a1852e9ff2e7cb61bd911cb964ae939e95621121a53b1e5af7c2cb341cd04283",
                "sha256:b59e6addd032a4e64c7a911245e7223688a0801b09bf41d3ff2979a4c0ad9249"
            ]
        },
        "Metadata": {
            "LastTagTime": "2020-08-25T20:21:43.247318778-04:00"
        }
    }
]
```

### (optional) gRPC Client/Server

(why not?)
```
docker run -p 50051:50051 bazel/greeter_server:greeter_server_image
docker run --net=host bazel/greeter_client:greeter_client_image
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

- `Client`
```bash
$ docker push salrashid123/greeter_server:greeter_server_image
    a1852e9ff2e7: Pushed 
    greeter_server_image: digest: sha256:e3e95e8f07b552ee2f60aaf6308b75ee660e24ff58d3a2b25be26f53476cde87 size: 738

```

On any other machine pull the image and inspect

```bash
$ docker inspect salrashid123/greeter_server@sha256:e3e95e8f07b552ee2f60aaf6308b75ee660e24ff58d3a2b25be26f53476cde87
[
    {
        "Id": "sha256:7ea0fc3e14c0d0cfdd8048f5ddd19566a1b78f822658b8c5318c14241340a982",
        "RepoTags": [
            "bazel/greeter_server:greeter_server_image",
            "salrashid123/greeter_server:greeter_server_image"
        ],
        "RepoDigests": [
            "salrashid123/greeter_server@sha256:e3e95e8f07b552ee2f60aaf6308b75ee660e24ff58d3a2b25be26f53476cde87"
        ],
   ...
```



### Cloud BUild

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

    Loaded image ID: sha256:7ea0fc3e14c0d0cfdd8048f5ddd19566a1b78f822658b8c5318c14241340a982
    Tagging 7ea0fc3e14c0d0cfdd8048f5ddd19566a1b78f822658b8c5318c14241340a982 as gcr.io/mineral-minutia-820/greeter_server:greeter_server_image
    PUSH
    Pushing gcr.io/mineral-minutia-820/greeter_server:greeter_server_image
    The push refers to repository [gcr.io/mineral-minutia-820/greeter_server]
    greeter_server_image: digest: sha256:e3e95e8f07b552ee2f60aaf6308b75ee660e24ff58d3a2b25be26f53476cde87 size: 738
    DONE
```

Then on the same system that it was built and pushed:

```bash
$ docker inspect bazel/greeter_server:greeter_server_image
[
    {
        "Id": "sha256:7ea0fc3e14c0d0cfdd8048f5ddd19566a1b78f822658b8c5318c14241340a982",
        "RepoTags": [
            "bazel/greeter_server:greeter_server_image",
            "salrashid123/greeter_server:greeter_server_image"
        ],
        "RepoDigests": [
            "salrashid123/greeter_server@sha256:e3e95e8f07b552ee2f60aaf6308b75ee660e24ff58d3a2b25be26f53476cde87"


```

### TODO:

- use `gazelle` for dependencies `bazel run //:gazelle -- update-repos -from_file=examples/greeter_server/go.mod`
