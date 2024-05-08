FROM docker.io/library/ubuntu:23.10 as base

ENV DEBIAN_FRONTEND=noninteractive

# RUN bash -c "yes || true" | unminimize

RUN apt-get update && apt-get install -yq \
ca-certificates \
curl \
build-essential \
git \
host \
vim \
man \
net-tools \
openssh-server \
iproute2 \
inetutils-ping \
sqlite3 \
postgresql \
jq \
git-lfs \
bash-completion \
&& rm -rf /var/lib/apt/lists/*

# pgadmin
RUN curl https://www.pgadmin.org/static/packages_pgadmin_org.pub | apt-key add
RUN echo "deb https://ftp.postgresql.org/pub/pgadmin/pgadmin4/apt/$(lsb_release -cs) pgadmin4 main" > /etc/apt/sources.list.d/pgadmin4.list

# Install Ubuntu packages
RUN apt-get update && apt-get install -yq \
pgadmin4-web \
&& rm -rf /var/lib/apt/lists/*

WORKDIR /opt/app

# renovate: depName=devsecops/dev-tools
ENV ACT3_DEV_TOOLS_VERSION=1.23.14
RUN --mount=type=secret,id=act3_token curl -H "PRIVATE-TOKEN: $(cat /run/secrets/act3_token)" -sSLfo /tmp/act3-dev-tools.deb https://git.act3-ace.com/api/v4/projects/365/packages/generic/dev-tools/${ACT3_DEV_TOOLS_VERSION}/act3-dev-tools.deb \
  && apt install --install-suggests --install-recommends /tmp/act3-dev-tools.deb \
  && rm -f /tmp/act3-dev-tools.deb


# Install tini
ENV TINI_VERSION v0.19.0
ADD https://github.com/krallin/tini/releases/download/${TINI_VERSION}/tini /tini
RUN chmod +x /tini
ENTRYPOINT ["/tini", "-g", "--"]

# Install VS Code Server
ENV CODE_VERSION=4.11.0
RUN curl -sSLo code-server.deb https://github.com/coder/code-server/releases/download/v${CODE_VERSION}/code-server_${CODE_VERSION}_amd64.deb \
&& apt-get install -y ./code-server.deb \
&& rm code-server.deb

# Install VS Code extensions to the server a list of extensions can be found on the VS Code extension marketplace, 
# https://marketplace.visualstudio.com/VSCode
RUN code-server --install-extension shd101wyy.markdown-preview-enhanced
RUN code-server --install-extension humao.rest-client
RUN code-server --install-extension eamodio.gitlens
RUN code-server --install-extension gitlab.gitlab-workflow
RUN code-server --install-extension alexcvzz.vscode-sqlite
RUN code-server --install-extension jq-syntax-highlighting.jq-syntax-highlighting

# Not working
# RUN code-server --install-extension davidanson.vscode-markdownlint
# RUN code-server --install-extension hediet.vscode-drawio
# RUN code-server --install-extension ms-kubernetes-tools.vscode-kubernetes-tools

# Why is bash completion not enabled by default when the bash-completion package is installed?
RUN echo ". /etc/bash_completion" >> ~/.bashrc

# Setup some defaults for VS Code Server
COPY settings.json /root/.local/share/code-server/User/settings.json

CMD ["code-server", "--auth", "none", "--port", "8888", "--host", "0.0.0.0"]

ENV GOLANG_VERSION=1.22.0
RUN curl -sSLo /tmp/go${GOLANG_VERSION}.linux-amd64.tar.gz https://golang.org/dl/go${GOLANG_VERSION}.linux-amd64.tar.gz \
&& rm -rf /usr/local/go \
&& tar -C /usr/local -xzf /tmp/go${GOLANG_VERSION}.linux-amd64.tar.gz \
&& rm /tmp/go${GOLANG_VERSION}.linux-amd64.tar.gz

ENV PATH $PATH:/usr/local/go/bin
# RUN mkdir -p /root/go
# ENV GOPATH /root/go

RUN code-server --install-extension golang.Go

RUN go install github.com/cweill/gotests/gotests@latest
RUN go install github.com/fatih/gomodifytags@latest

RUN go install github.com/uudashr/gopkgs/v2/cmd/gopkgs@latest
RUN go install github.com/ramya-rao-a/go-outline@latest
RUN go install github.com/josharian/impl@latest
RUN go install github.com/haya14busa/goplay/cmd/goplay@latest
RUN go install github.com/go-delve/delve/cmd/dlv@latest
RUN go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
RUN go install golang.org/x/tools/gopls@latest

RUN go env -w GOPRIVATE=git.act3-ace.com
