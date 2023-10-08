import pytest
from . import helpers


@pytest.mark.order(20)
def test_providers():
    helpers.assert_xdeb_install_command("providers")


@pytest.mark.order(21)
def test_providers_details():
    helpers.assert_xdeb_install_command("providers", "--details")
