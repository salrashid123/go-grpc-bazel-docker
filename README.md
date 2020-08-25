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
bazel build  --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 greeter_client:all

bazel run  --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 greeter_client:greeter_client_image
bazel run  --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 greeter_server:greeter_server_image
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
$ docker inspect bazel/greeter_client:greeter_client_image
[
    {
        "Id": "sha256:c9197e9819cdac12a89abcd1158b0bdf78a1d23d50796e4c73d6530079133606",
        "RepoTags": [
            "bazel/greeter_client:greeter_client_image"
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
        "Size": 15759806,
        "VirtualSize": 15759806,
        "GraphDriver": {
            "Data": {
                "LowerDir": "/var/lib/docker/overlay2/066c5b6ddf94a3b31f953b94453f43d5884afa06e2d8bd15a9d073685556c9e4/diff",
                "MergedDir": "/var/lib/docker/overlay2/1bf40026b95985b478ecf35a15e9a01678fefd4acc180b722b9682430f370443/merged",
                "UpperDir": "/var/lib/docker/overlay2/1bf40026b95985b478ecf35a15e9a01678fefd4acc180b722b9682430f370443/diff",
                "WorkDir": "/var/lib/docker/overlay2/1bf40026b95985b478ecf35a15e9a01678fefd4acc180b722b9682430f370443/work"
            },
            "Name": "overlay2"
        },
        "RootFS": {
            "Type": "layers",
            "Layers": [
                "sha256:7444ea29e45e927abea1f923bf24cac20deaddea603c4bb1c7f2f5819773d453",
                "sha256:abd61d679f7f6dfe3426e1a652427082e392715ed9683e0d7f84b472e462f104"
            ]
        },
        "Metadata": {
            "LastTagTime": "2020-08-24T20:09:51.803295418-04:00"
        }
    }
]
```

### (optional) gRPC Client/Server
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

```bash
$ docker inspect bazel/greeter_client:greeter_client_image
[
    {
        "Id": "sha256:c9197e9819cdac12a89abcd1158b0bdf78a1d23d50796e4c73d6530079133606",
        "RepoTags": [
            "bazel/greeter_client:greeter_client_image",
            "salrashid123/greeter_client:greeter_client_image"
        ],
        "RepoDigests": [],
        "Parent": "",
        "Comment": "",
        "Created": "1970-01-01T00:00:00Z",
        "Container": "e61c47f59d8323f7c6db62b1c47bb70faf0d8604756b85e9eb0cd329e88872d8",

```

on push to dockerhub

- `Client`
```bash
$ docker push salrashid123/greeter_client:greeter_client_image
    The push refers to repository [docker.io/salrashid123/greeter_client]
    greeter_client_image: digest: sha256:55746694a16db7a2036984088b3198ffae3ad9d90f7f05208c3c5d6b9e64a633 size: 738

$ docker inspect bazel/greeter_client:greeter_client_image
[
    {
        "Id": "sha256:c9197e9819cdac12a89abcd1158b0bdf78a1d23d50796e4c73d6530079133606",
        "RepoTags": [
            "bazel/greeter_client:greeter_client_image",
            "salrashid123/greeter_client:greeter_client_image"
        ],
        "RepoDigests": [
            "salrashid123/greeter_client@sha256:55746694a16db7a2036984088b3198ffae3ad9d90f7f05208c3c5d6b9e64a633"
        ],
        "Parent": "",
        "Comment": "",
        "Created": "1970-01-01T00:00:00Z",
        "Container": "e61c47f59d8323f7c6db62b1c47bb70faf0d8604756b85e9eb0cd329e88872d8",
```

- `Server`
```bash
$  docker push salrashid123/greeter_server:greeter_server_image
    The push refers to repository [docker.io/salrashid123/greeter_server]
    greeter_server_image: digest: sha256:ccd3f4776ff236f7455281c74da39c2d5d9cdc5a9ad31f75b2bc38773539fef3 size: 738

$ docker inspect bazel/greeter_server:greeter_server_image
[
    {
        "Id": "sha256:b26b67e46ab3732c542b93b3988cd4419cc3e9d137b85654b9b64117f4fe8e43",
        "RepoTags": [
            "bazel/greeter_server:greeter_server_image",
            "salrashid123/greeter_server:greeter_server_image"
        ],
        "RepoDigests": [
            "salrashid123/greeter_server@sha256:ccd3f4776ff236f7455281c74da39c2d5d9cdc5a9ad31f75b2bc38773539fef3"
        ],
        "Parent": "",
        "Comment": "",
        "Created": "1970-01-01T00:00:00Z",
        "Container": "e61c47f59d8323f7c6db62b1c47bb70faf0d8604756b85e9eb0cd329e88872d8",    
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

The push refers to repository [gcr.io/mineral-minutia-820/greeter_server]
b59e6addd032: Preparing
7444ea29e45e: Preparing
7444ea29e45e: Pushed
b59e6addd032: Pushed
greeter_server_image: digest: sha256:ccd3f4776ff236f7455281c74da39c2d5d9cdc5a9ad31f75b2bc38773539fef3 size: 738
DONE
---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

ID                                    CREATE_TIME                DURATION  SOURCE                                                                                             IMAGES                                                          STATUS
d3ec4b0c-3135-4d2e-ac58-3a34d8d8a5e6  2020-08-25T00:41:50+00:00  1M51S     gs://mineral-minutia-820_cloudbuild/source/1598316109.957319-67b9f7f8148d496287bfecb45a09fa97.tgz  gcr.io/mineral-minutia-820/greeter_server:greeter_server_image  SUCCESS
```
```bash
$ docker inspect bazel/greeter_server:greeter_server_image
[
    {
        "Id": "sha256:b26b67e46ab3732c542b93b3988cd4419cc3e9d137b85654b9b64117f4fe8e43",
        "RepoTags": [
            "bazel/greeter_server:greeter_server_image",
            "salrashid123/greeter_server:greeter_server_image",
            "gcr.io/mineral-minutia-820/greeter_server:greeter_server_image"
        ],
        "RepoDigests": [
            "salrashid123/greeter_server@sha256:ccd3f4776ff236f7455281c74da39c2d5d9cdc5a9ad31f75b2bc38773539fef3",
            "gcr.io/mineral-minutia-820/greeter_server@sha256:ccd3f4776ff236f7455281c74da39c2d5d9cdc5a9ad31f75b2bc38773539fef3"
        ],

```

### TODO:

- use `gazelle` for dependencies `bazel run //:gazelle -- update-repos -from_file=examples/greeter_server/go.mod`
