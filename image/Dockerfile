FROM mcr.microsoft.com/vscode/devcontainers/base:debian-12 AS microsoft_devcontainer_image
FROM ghcr.io/void-linux/void-glibc-full:20231003R1

RUN xbps-install -Syu \
    && xbps-install -y \
        git \
        tar \
        util-linux \
        grep \
        curl \
        shadow \
        vim \
        nano \
        go \
        sudo \
        openssh \
        bash \
        binutils \
        xz \
        zstd \
        wget \
        ncurses \
        readline \
        bash-completion \
        python3 \
        python3-pip \
        python3-pytest \
    && xbps-remove -yOo \
    && chsh -s /bin/bash \
    && groupadd -g 1000 user \
    && useradd -g user -ms /bin/bash -u 1000 user \
    && echo "user ALL=(ALL:ALL) NOPASSWD: ALL" > /etc/sudoers.d/user

USER user
WORKDIR /home/user

COPY --from=microsoft_devcontainer_image /root/.bashrc /root/.bashrc
COPY --chown=user:user --from=microsoft_devcontainer_image /home/vscode/.bashrc /home/user/.bashrc

RUN go install golang.org/x/tools/gopls@latest \
    && go install github.com/go-delve/delve/cmd/dlv@latest \
    && go install honnef.co/go/tools/cmd/staticcheck@latest \
    && pip3 install --user \
        pytest-order \
        pytest-html

ENV GOPATH=/home/user/go
ENV PATH=${PATH}:${GOPATH}/bin
