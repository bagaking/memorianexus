# Memoria Nexus

Memoria Nexus is a Go backend prototype for memory-assisted study workflows.
The current repository is not a complete web learning product: it contains a
Go module, API/service scaffolding, database migrations, generated API docs,
and tested memory-curve calculation code.

## Current Status

Implemented and verifiable today:

- Go module: `github.com/bagaking/memorianexus`
- HTTP backend entrypoint using Gin, Gorm, MySQL, Redis, and Swagger wiring
- Route scaffolding for profiles, items, books, tags, dungeons, campaigns,
  analytics, NFTs, achievements, operations, and system endpoints
- Memory-curve review calculation package in `pkg/memcurve` with unit tests
- Tag package tests in `pkg/tags`
- Makefile target `make test`, which runs `go test ./...`
- GitHub Actions workflow `.github/workflows/test.yml`, which runs `make test`
  on pushes to `main`/`master` and on pull requests
- Docker-oriented development and deployment assets, including migrations and a
  Dockerfile

Prototype or scaffolded areas:

- Review session, reminder, and core analytics packages under `src/core/` are
  planned extensions with placeholder implementations, not production-complete
  features.
- The analytic module currently registers handlers, but those handlers do
  not yet return analytics data.
- Several gamification and NFT-facing modules expose route shapes before full
  business behavior is implemented.
- No React, Redux, Sass, or other frontend application is present in this
  repository.
- Local runtime setup still depends on external services and development
  configuration; the most reliable lightweight validation path is the Go test
  suite.

## Project Layout

See [Project Structure](./doc/CODE_STRUCTURE.md) for the current code layout.
Generated Swagger output is available under `doc/`.

## Local Validation

Run the Go test suite before opening a pull request:

```sh
make test
```

This target runs:

```sh
go test ./...
```

## Development Notes

- Use the Makefile for common local tasks.
- Use the Docker and migration files under `deployment/` and
  `dev_memorianexus/` when working on service-level runtime setup.
- Treat user-facing study flows, reminders, review sessions, analytics
  dashboards, and frontend UI as unfinished until their handlers and tests are
  implemented.

## License

Memoria Nexus is licensed under the MIT License. See [LICENSE](LICENSE) for
more information.
