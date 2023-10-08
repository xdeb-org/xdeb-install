import subprocess

from . import constants


def assert_xdeb_install_command(*args):
    subprocess.check_call([constants.XDEB_INSTALL_BINARY_PATH, *args])


def assert_command_assume_yes(returncode: int, *args):
    process = subprocess.run(*args, input="yes\n".encode(), stdout=subprocess.PIPE)
    assert process.returncode == returncode


def assert_xdeb_install_xbps(returncode: int, *args):
    assert_command_assume_yes(returncode, [constants.XDEB_INSTALL_BINARY_PATH, *args])

    if returncode == 0:
        package = args[-1] if args[-1] not in constants.XDEB_INSTALL_PACKAGE_MAP else constants.XDEB_INSTALL_PACKAGE_MAP[args[-1]]
        assert_command_assume_yes(0, ["sudo", "xbps-remove", package])
        assert_command_assume_yes(0, ["sudo", "xbps-remove", "-Oo"])


def assert_xdeb_installation(version: str = None):
    # remove any previous xdeb binary
    subprocess.check_call(["sudo", "rm", "-rf", constants.XDEB_BINARY_PATH])

    # install xdeb version
    args = ["xdeb"]

    if version is not None:
        args.append(version)

    assert_xdeb_install_command(*args)
    subprocess.check_call([constants.XDEB_BINARY_PATH, "-h"])
