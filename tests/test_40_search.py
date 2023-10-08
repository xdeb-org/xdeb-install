import subprocess
import pytest

from . import constants
from . import helpers


@pytest.mark.order(40)
def test_search_nothing():
    with pytest.raises(subprocess.CalledProcessError):
        helpers.assert_xdeb_install_command("search")


@pytest.mark.order(41)
def test_search_nonexistent():
    with pytest.raises(subprocess.CalledProcessError):
        helpers.assert_xdeb_install_command("search", constants.DEB_NONEXISTENT_PACKAGE)


@pytest.mark.order(42)
def test_search_speedcrunch():
    helpers.assert_xdeb_install_command("search", "speedcrunch")


@pytest.mark.order(43)
def test_search_packages():
    for package, provider_data in constants.XDEB_INSTALL_HAVE_PACKAGE.items():
        for provider, data in provider_data.items():
            if not data["any"]:
                with pytest.raises(subprocess.CalledProcessError):
                    helpers.assert_xdeb_install_command("search", "--provider", provider, package)
                continue

            helpers.assert_xdeb_install_command("search", "--provider", provider, package)

            for distribution, available in data["distributions"].items():
                if not available:
                    with pytest.raises(subprocess.CalledProcessError):
                        helpers.assert_xdeb_install_command("search", "--provider", provider, "--distribution", distribution, package)
                    continue

                helpers.assert_xdeb_install_command("search", "--provider", provider, "--distribution", distribution, package)
