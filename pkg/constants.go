package xdeb

import "time"

const APPLICATION_NAME = "xdeb-install"

const LOG_MESSAGE_PREFIX = "[" + APPLICATION_NAME + "]"

const XDEB_URL = "https://github.com/xdeb-org/xdeb/releases"
const XDEB_MASTER_URL = "https://raw.githubusercontent.com/xdeb-org/xdeb/master"

const XDEB_INSTALL_REPOSITORIES_TAG = "v1.1.2"
const XDEB_INSTALL_REPOSITORIES_URL = "https://raw.githubusercontent.com/xdeb-org/xdeb-install-repositories"

const HTTP_REQUEST_HEADERS_TIMEOUT = 10 * time.Second
