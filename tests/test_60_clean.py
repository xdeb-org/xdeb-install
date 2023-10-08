import pytest

from . import helpers


@pytest.mark.order(60)
def test_clean():
    helpers.assert_xdeb_install_command("clean")


@pytest.mark.order(61)
def test_clean_lists():
    helpers.assert_xdeb_install_command("clean", "--lists")
