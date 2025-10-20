# ---------------------------------------------------------#
#                   Build harness-migrator image           #
# ---------------------------------------------------------#
FROM golang:1.23-alpine

# install required dependencies
RUN apk update && \
    apk add --no-cache \
    git \
    git-lfs \
    ca-certificates \
    tzdata \
    bash \
    curl \
    openssh \
    build-base

# install git-lfs without repo-level hooks
RUN git lfs install --skip-repo

# set working directory
WORKDIR /app

# copy go mod and sum files
COPY go.mod go.sum ./

# download dependencies
RUN go mod download

COPY . .

# set required build flags
ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# build the application
RUN go build -o harness-migrate main.go

# set git config for the container
RUN git config --system credential.helper cache && \
    git config --system http.sslVerify true && \
    git config --system core.longpaths true && \
    git config --system core.compression 9

# create non-root user and set permissions
RUN adduser -D -u 1001 harness-migrate && \
    mkdir -p /data && \
    chown -R harness-migrate:harness-migrate /app /data && \
    chmod 755 /app

# switch to non-root user
USER harness-migrate

# add harness-migrate to PATH
ENV PATH="/app:${PATH}"

# set the entrypoint
ENTRYPOINT ["/bin/bash"]
