# YouThumb

gRPC service for downloading thumbnail images from YouTube.

## gRPC API

The user sends a request to get a thumbnail image by the URL of a YouTube video.
The service returns the image as a sequence of chunks.

```proto
package youthumb.v1;

service ThumbnailService {
  rpc GetThumbnail(GetThumbnailRequest) returns (stream ThumbnailChunk);
}

message GetThumbnailRequest {
  string video_url = 1;
}

message ThumbnailChunk {
  string content_type = 1;
  bytes data = 2;
}
```

You can use both regular and short URLs as `video_url`.
For example, the links `https://www.youtube.com/watch?v=dQw4w9WgXcQ` and `https://youtu.be/dQw4w9WgXcQ` are equivalent.
More supported formats can be seen in the test [`internal/thumbnail/url_test.go`](internal/thumbnail/url_test.go).

The full service definition and documentation are in the file [`proto/youthumb/v1/youthumb.proto`](proto/youthumb/v1/youthumb.proto).

## Architecture

Two entry points:

- [`cmd/server`](cmd/server) - entry point for running the gRPC server.
- [`cmd/client`](cmd/client) - example gRPC client for sending requests to the server.

Main package with business logic:

- [`internal/thumbnail`](internal/thumbnail) - package with the service's business logic and gRPC server implementation.

Auxiliary packages for gRPC:

- [`internal/rpc`](internal/rpc) - package with the main gRPC server constructor.
- [`internal/rpc/interceptor`](internal/rpc/interceptor) - package with middleware for the gRPC server.
- [`internal/rpc/message`](internal/rpc/message) - package with common messages for the gRPC server.

Auxiliary packages for the application:

- [`internal/app`](internal/app) - package with common application components.
- [`internal/app/config`](internal/app/config) - package with application configuration.
- [`internal/app/log`](internal/app/log) - package with the application logger.

## Installation and Running

> *Note:* For testing, the repository contains files with video links. The files are located in the [`examples`](examples) folder.
> The files with 50 and 250 links for testing contain links to non-existent videos to demonstrate error handling.

### Manually

Running the server with an SQLite database for caching:

```sh
$ go run ./cmd/server -d db.sqlite3
```

Running the client to download images for 50 YouTube videos asynchronously, results are saved in the `./results` folder:

```sh
$ go run ./cmd/client -async -o ./results ./examples/video_urls_50.txt
```

### Docker Compose

> *Warning:* Inside the containers, a regular user `user` is used, so when running the containers,
> you need to pass the current user to the containers to avoid file access permission issues.

Building and running the server with an SQLite database for caching:

```sh
$ docker compose up --build
```

Building and running the client to download images for 50 YouTube videos asynchronously, results are saved in the `./results` folder:

```sh
# This docker compose run command:
# - Makes the current working directory available inside the Docker container (--volume).
# - Forces the client to use the same user inside the container so it can access the working directory (--user).
# - Helps the client avoid issues if *your* user does not exist inside the container by setting HOME to a *writable* directory (--env).
$ docker compose run --build \
    --volume "$(pwd):/user/data" \
    --user "$(id -u):$(id -g)" \
    --env HOME=/tmp \
    client -async -o ./results examples/video_urls_50.txt
```

## Demo

[![](assets/demo.png)](https://drive.google.com/file/d/18OGnqKGRguiHuV0eoTHgOyJAUHd66tS6/view?usp=sharing)

In the [video](https://drive.google.com/file/d/18OGnqKGRguiHuV0eoTHgOyJAUHd66tS6/view?usp=sharing), the following is shown:

1. Building the client and server.
2. Running the server in the right terminal on a clean database.
3. Running the client in the left terminal to download images for 250 YouTube videos asynchronously. The client correctly outputs messages about three not found videos with IDs `00000000001`, `00000000002`, and `00000000003`.
4. Showing the contents of the `./results` folder with the downloaded images.
5. Deleting the `./results` folder and re-running the client for the same videos. This time, the client gets images from the server cache, as seen by the noticeably faster completion of the client's work.
6. Re-showing the contents of the `./results` folder with the downloaded images.
