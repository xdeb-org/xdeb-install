import pytest

from . import constants
from . import helpers


@pytest.mark.order(10)
def test_xdeb():
    helpers.assert_xdeb_installation()

    for release in constants.XDEB_RELEASES:
        helpers.assert_xdeb_installation(release)

    helpers.assert_xdeb_installation("latest")
