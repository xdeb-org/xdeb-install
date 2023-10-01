package xdeb

import "fmt"

func LogMessage(format string, args ...any) {
	message := fmt.Sprintf(format, args...)
	fmt.Printf("%s %s\n", LOG_MESSAGE_PREFIX, message)
}
