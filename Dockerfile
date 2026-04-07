FROM golang:1.25-alpine AS builder

ARG CHECKER_VERSION=custom-build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags "-X main.Version=${CHECKER_VERSION}" -o /checker-matrix .

FROM scratch
COPY --from=builder /checker-matrix /checker-matrix
EXPOSE 8080
ENTRYPOINT ["/checker-matrix"]
