<p align="center">
  <a href="https://github.com/mateusfdl/go-api/actions/workflows/commit.yaml" target="_blank"><img src="https://github.com/mateusfdl/go-api/actions/workflows/commit.yaml/badge.svg" alt="Commit Pipeline" /></a>
  <a href="https://github.com/mateusfdl/go-api/actions/workflows/build.yaml" target="_blank"><img src="https://github.com/mateusfdl/go-api/actions/workflows/build.yaml/badge.svg" alt="Build Pipeline" /></a>
</p>

# Go API

## Installation

You can run the application by cloning this repository or by pulling the Docker image with the desired tag. Using the `latest` tag is recommended.

```bash
$ docker pull ghcr.io/mateusfdl/go-api:latest
```

- If cloning this repository and using Docker is not an option, you can download one of the releases from the [releases page](https://github.com/mateusfdl/go-api/releases).

<br>
<div style="display: inline-block; background-color: rgba(255, 255, 0, 0.5); border: 2px solid rgb(255, 204, 0); padding: 10px; border-radius: 5px;">
⚠️ WARNING: If you are not using Docker, you must have a MongoDB instance up and running locally or remotely.
</div>

## Running the App Locally

### Requirements

Ensure the following dependencies are installed:

- [Docker](https://docs.docker.com/get-docker/)
- [Docker Compose](https://docs.docker.com/compose/)
- [MongoDB](https://www.mongodb.com/docs/manual/installation/)

Start MongoDB using Docker Compose:

```bash
$ docker compose up -d mongo
```

### Configuration

Set up your environment variables using the provided `.env-example` file. To quickly set up a local environment, run:

```bash
$ cp .env-example .env
```

Update the `.env` file with your specific configuration as needed.

### Running Locally

Even though It's a straightforward API, still some deps are needed. You can download them by

```bash
$ go mod tidy
```

Start the API:

```bash
$ go run ./cmd/app/main.go
```

Check the health of the application with the following command:

```bash
$ curl 'localhost:3000/health'
```

## CI

### Linting

Run static analysis and formatting checks on your code using the following commands:

```bash
$ go vet ./...
$ go fmt ./...
```

Alternatively, you can use [GolangCI-Lint](https://github.com/golangci/golangci-lint), which is also used in the CI pipeline:

```bash
$ golangci-lint run ./...
```

You can also run GolangCI-Lint using Docker:

```bash
$ docker run --rm -v $(pwd):/app -w /app golangci/golangci-lint:v1.62.0 golangci-lint run -v
```

### Running Unit Tests

Unit tests are located alongside the packages they test. Run them using:

```bash
$ go test -v ./internal/... ./config/...
```

### Running Integration Tests

Integration tests are located in the `tests/` directory. Run them using:

```bash
$ go test -v ./test/...
```

## API

### Specs

- You can import the `openapi.yaml` file into your preferred API tool (e.g., Postman, Swagger UI) as a reference for the current API specifications.
- Alternatively, you can run the docker swagger image by just

```bash
$ docker compose up -d swagger-ui
```

- It will bind to the port 8080, so you can hit it through your browser. Just make sure the API is running locally or in the container

### Logging

- Toggle between sugar logging and standard logging by setting the `LOGGER_SUGARED` variable in your `.env` file to `false`.
- Change the log level by modifying the `LOG_LEVEL` variable. For example, setting `LOG_LEVEL=error` will only log errors to `STDOUT`.

<br><br><br><br>
<h1 align="center"> Happy Hacking :)</h1>

