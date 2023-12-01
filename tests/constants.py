from pathlib import Path


DEB_NONEXISTENT_PACKAGE = "DEFINITELY-NON-EXISTENT-PACKAGE"


XDEB_BINARY_PATH = Path("/usr/local/bin/xdeb")
XDEB_RELEASES = (
    "1.0",
    "1.1",
    "1.2",
    "1.3"
)

XDEB_INSTALL_BINARY_PATH = Path(__file__).parent.parent.joinpath("bin", "xdeb-install-linux-x86_64")

XDEB_INSTALL_PROVIDERS = (
    "debian.org",
    "linuxmint.com",
    "ubuntu.com",
    "microsoft.com",
    "google.com",
)

XDEB_INSTALL_HAVE_PACKAGE = {
    "speedcrunch": {
        "debian.org": {
            "any": True,
            "distributions": {
                "bookworm": True,
                "bookworm-backports": False,
                "bullseye": True,
                "bullseye-backports": False,
                "buster": True,
                "buster-backports": False,
                "sid": True,
                "testing": False,
                "testing-backports": False
            },
        },
        "linuxmint.com": {
            "any": False
        },
        "ubuntu.com": {
            "any": True,
            "distributions": {
                "bionic": True,
                "focal": True,
                "jammy": True
            }
        },
        "microsoft.com": {
            "any": False
        },
        "google.com": {
            "any": False
        },
    },
    "vscode": {
        "debian.org": {
            "any": False
        },
        "linuxmint.com": {
            "any": False
        },
        "ubuntu.com": {
            "any": False
        },
        "microsoft.com": {
            "any": True,
            "distributions": {
                "current": True
            }
        },
        "google.com": {
            "any": False
        }
    },
    "google-chrome": {
        "debian.org": {
            "any": False
        },
        "linuxmint.com": {
            "any": False
        },
        "ubuntu.com": {
            "any": False
        },
        "microsoft.com": {
            "any": False
        },
        "google.com": {
            "any": True,
            "distributions": {
                "current": True
            }
        }
    },
    "google-chrome-unstable": {
        "debian.org": {
            "any": False
        },
        "linuxmint.com": {
            "any": False
        },
        "ubuntu.com": {
            "any": False
        },
        "microsoft.com": {
            "any": False
        },
        "google.com": {
            "any": True,
            "distributions": {
                "current": True
            }
        }
    }
}

XDEB_INSTALL_PACKAGE_MAP = {
    "vscode": "code",
    "google-chrome": "google-chrome-stable"
}
