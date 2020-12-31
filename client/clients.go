package client

import "fmt"

func wrapError(cumstomMsg string, originalError error) error {
	return fmt.Errorf("%s : %v", cumstomMsg, originalError)
}
