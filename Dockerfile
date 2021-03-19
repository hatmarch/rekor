FROM golang:1.16.2 AS builder
ENV APP_ROOT=/opt/app-root
ENV GOPATH=$APP_ROOT

WORKDIR $APP_ROOT/src/
ADD go.mod go.sum $APP_ROOT/src/
RUN go mod download

# Add source code
ADD ./ $APP_ROOT/src/

RUN go build ./cmd/server
RUN CGO_ENABLED=0 go build -gcflags "all=-N -l" -o server_debug ./cmd/server

# Multi-Stage production build
FROM golang:1.16.2 as deploy

# Retrieve the binary from the previous stage
COPY --from=builder /opt/app-root/src/server /usr/local/bin/rekor-server

# Set the binary as the entrypoint of the container
CMD ["rekor-server", "serve"]

# debug compile options & debugger
FROM deploy as debug
RUN go get github.com/go-delve/delve/cmd/dlv

# overwrite server and include debugger
COPY --from=builder /opt/app-root/src/server_debug /usr/local/bin/rekor-server
RUN mkdir -p /opt/app-root/src && chown -R 0:0 /opt/app-root && chmod -R 775 /opt/app-root
COPY --from=builder --chown=0:0 --chmod=775 /opt/app-root/src/ /opt/app-root/src
#RUN chmod -R g+w /opt/app-root/ && chmod -R g+w /go
ENV HOME=/opt/app-root
