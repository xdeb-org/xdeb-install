import pytest

from . import helpers


@pytest.mark.order(0)
def test_version():
    helpers.assert_xdeb_install_command("--version")


@pytest.mark.order(1)
def test_help():
    helpers.assert_xdeb_install_command("--help")
