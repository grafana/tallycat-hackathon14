# syntax=docker/dockerfile:1.3
# Dockerfile reference: https://docs.docker.com/engine/reference/builder/

FROM --platform=${BUILDPLATFORM} node:24 AS ui-builder
WORKDIR /tallycat
COPY ui /tallycat
RUN npm ci
RUN npm run build

FROM --platform=${TARGETPLATFORM} golang:1.24-trixie AS builder

WORKDIR /tallycat

COPY go.mod /tallycat
COPY go.sum /tallycat

RUN go mod download

COPY main.go /tallycat
COPY internal /tallycat/internal
COPY cmd /tallycat/cmd
COPY ui/embed.go /tallycat/ui/embed.go
COPY --from=ui-builder /tallycat/dist /tallycat/ui/dist

RUN CGO_ENABLED=1 go build -o tallycat ./

FROM --platform=${TARGETPLATFORM} debian:trixie-slim as executor

USER 1000:1000
WORKDIR /tallycat

COPY --chown=1000:1000 --from=builder /tallycat/tallycat .
COPY --from=otel/weaver:v0.18.0 /weaver/weaver /usr/bin/weaver
COPY weaver/templates /opt/weaver/templates

EXPOSE 8080
EXPOSE 4317
ENTRYPOINT [ "/tallycat/tallycat" ]
CMD [ "server" ]
