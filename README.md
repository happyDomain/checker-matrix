# checker-matrix

Matrix federation checker for [happyDomain](https://www.happydomain.org/).

Queries a [Matrix Federation Tester](https://federationtester.matrix.org/)
instance to verify that a Matrix homeserver is correctly federating, stores
the full report as an observation, and renders a rich HTML summary
(connections, certificates, well-known, DNS/SRV resolution).

## Usage

### Standalone HTTP server

```bash
make
./checker-matrix -listen :8080
```

The server exposes the standard happyDomain external checker endpoints
(`/health`, `/definition`, `/collect`, `/evaluate`, `/html-report`).

### Docker

```bash
make docker
docker run -p 8080:8080 happydomain/checker-matrix
```

### happyDomain plugin

```bash
make plugin
# produces checker-matrix.so, loadable by happyDomain as a Go plugin
```

The plugin exposes a `NewCheckerPlugin` symbol returning the checker
definition and observation provider, which happyDomain registers in its
global registries at load time.

### Versioning

The binary, plugin, and Docker image embed a version string overridable
at build time:

```bash
make CHECKER_VERSION=1.2.3
make plugin CHECKER_VERSION=1.2.3
make docker CHECKER_VERSION=1.2.3
```

### happyDomain remote endpoint

Set the `endpoint` admin option for the `matrixim` checker to the URL of
the running checker-matrix server (e.g., `http://checker-matrix:8080`).
happyDomain will delegate observation collection to this endpoint.

## Rules

| Code                         | Description                                                                                       | Severity            |
|------------------------------|---------------------------------------------------------------------------------------------------|---------------------|
| `matrix.connection_reachable`| Checks that every discovered federation endpoint accepts an inbound connection.                   | CRITICAL            |
| `matrix.federation_ok`       | Reports the overall federation status returned by the Matrix Federation Tester.                   | CRITICAL            |
| `matrix.srv_records`         | Checks that the Matrix SRV lookup (`_matrix-fed._tcp` / `_matrix._tcp`) succeeded or was skipped. | CRITICAL            |
| `matrix.tls_checks`          | Reviews the TLS posture on every reachable federation endpoint (chain, hostname, Ed25519 key).    | CRITICAL            |
| `matrix.version`             | Checks that the homeserver responds to `/_matrix/federation/v1/version` with name and version.    | WARNING             |
| `matrix.well_known`          | Checks that `/.well-known/matrix/server` (if published) is valid and points at the server_name.   | CRITICAL            |

## Options

| Scope | Id                       | Description                                                                |
| ----- | ------------------------ | -------------------------------------------------------------------------- |
| Run   | `serviceDomain`          | Matrix domain to test (auto-filled, default `matrix.org`)                  |
| Admin | `federationTesterServer` | Federation Tester URL template (default: `https://federationtester.matrix.org/api/report?server_name=%s`) |

The checker only applies to services of type `abstract.MatrixIM`.

## License

This project is licensed under the **MIT License** (see `LICENSE`). The
third-party Apache-2.0 attributions for `checker-sdk-go` are recorded in
`NOTICE` and must accompany any binary or source redistribution of this
project.
