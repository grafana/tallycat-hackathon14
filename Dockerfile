
FROM ubuntu:22.04


ENV DEBIAN_FRONTEND=noninteractive
ENV GO_VERSION=1.21.5
ENV NODE_VERSION=20


RUN apt-get update && apt-get install -y \
    curl \
    wget \
    build-essential \
    ca-certificates \
    git \
    libc6-dev \
    gcc \
    g++ \
    libstdc++6 \
    libgcc-s1 \
    && rm -rf /var/lib/apt/lists/*

RUN wget https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz \
    && tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz \
    && rm go${GO_VERSION}.linux-amd64.tar.gz


RUN curl -fsSL https://deb.nodesource.com/setup_${NODE_VERSION}.x | bash - \
    && apt-get install -y nodejs


ENV PATH=/usr/local/go/bin:$PATH
ENV GOPATH=/go
ENV PATH=$GOPATH/bin:$PATH
ENV CGO_ENABLED=1

WORKDIR /app


COPY go.mod go.sum ./

RUN go mod download


COPY . .
WORKDIR /app/ui
RUN npm install


WORKDIR /app

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

EXPOSE 4000 8080


ENTRYPOINT ["/entrypoint.sh"]