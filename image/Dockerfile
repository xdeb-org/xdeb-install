FROM ghcr.io/void-linux/void-glibc-full:20231202R1

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
        ncurses-base \
        bash-completion \
        python3 \
        python3-pip \
        python3-pytest \
    && xbps-remove -yOo \
    && chsh -s /bin/bash \
    && go install golang.org/x/tools/gopls@latest \
    && go install github.com/go-delve/delve/cmd/dlv@latest \
    && go install honnef.co/go/tools/cmd/staticcheck@latest \
    && pip3 install \
        pytest-order \
        pytest-html

ENV GOPATH=/root/go
ENV PATH=${PATH}:${GOPATH}/bin
