import pytest

from . import constants
from . import helpers


@pytest.mark.order(51)
def test_install_nothing():
    helpers.assert_xdeb_install_xbps(1)


@pytest.mark.order(52)
def test_install_nonexistent():
    helpers.assert_xdeb_install_xbps(1, constants.DEB_NONEXISTENT_PACKAGE)


@pytest.mark.order(53)
def test_install_speedcrunch():
    helpers.assert_xdeb_install_xbps(0, "speedcrunch")


@pytest.mark.order(54)
def test_install_packages():
    for package, provider_data in constants.XDEB_INSTALL_HAVE_PACKAGE.items():
        for provider, data in provider_data.items():
            if not data["any"]:
                helpers.assert_xdeb_install_xbps(1, "--provider", provider, package)
                continue

            helpers.assert_xdeb_install_xbps(0, "--provider", provider, package)

            for distribution, available in data["distributions"].items():
                if not available:
                    helpers.assert_xdeb_install_xbps(1, "--provider", provider, "--distribution", distribution, package)
                    continue

                helpers.assert_xdeb_install_xbps(0, "--provider", provider, "--distribution", distribution, package)
