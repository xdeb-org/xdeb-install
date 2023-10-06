FROM mcr.microsoft.com/vscode/devcontainers/base:debian-12 AS microsoft_devcontainer_image
FROM ghcr.io/thetredev/voidlinux-repository:x86_64

RUN xbps-install -Sy \
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
    && go install honnef.co/go/tools/cmd/staticcheck@latest
