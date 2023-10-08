import pytest

from . import constants
from . import helpers


@pytest.mark.order(30)
def test_sync():
    helpers.assert_xdeb_install_command("sync")


@pytest.mark.order(31)
def test_sync_single():
    for provider in constants.XDEB_INSTALL_PROVIDERS:
        helpers.assert_xdeb_install_command("sync", provider)


@pytest.mark.order(32)
def test_sync_each():
    helpers.assert_xdeb_install_command("sync", *constants.XDEB_INSTALL_PROVIDERS)
