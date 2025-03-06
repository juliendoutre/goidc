# goidc

A simple CLI tool written in Go to quickly get OIDC JWTs from various providers.

## Getting started

```shell
brew tap juliendoutre/goidc https://github.com/juliendoutre/goidc
brew install goidc
```

## Usage

```shell
goidc -help
goidc -version
 CLIENT_ID=... CLIENT_SECRET=... goidc
 CLIENT_ID=... CLIENT_SECRET=... goidc -decode
```

## Development

### Lint the code

```shell
brew install golangci-lint hadolint
golangci-lint run
hadolint ./Dockerfile
```

### Release a new version

```shell
git tag -a v0.1.0 -m "New release"
git push origin v0.1.0
```
