package printer

import (
	"fmt"
	"io"
)

func RateLimitWarning(w io.Writer) {
	fmt.Fprintf(w, "\U000026A0  You have reached the rate limit for the GitHub API. These results may not be accurate.\n\n")
}
