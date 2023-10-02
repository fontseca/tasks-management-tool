package failure

import (
	"fmt"

	"github.com/lib/pq"
)

func PQErrorToString(err *pq.Error) string {
	if err.Hint == "" {
		err.Hint = "(none)"
	}
	if err.Detail == "" {
		err.Detail = "(none)"
	}
	return fmt.Sprintf("postgres driver failed with error \033[1;31m%s\033[0m (%s):\n"+
		"  message: \033[0;33m%s\033[0m\n"+
		"   detail: %s\n"+
		"     hint: %s",
		err.Code, err.Code.Name(), err.Message, err.Detail, err.Hint)
}
