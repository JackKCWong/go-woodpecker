package ossindex

import (
	"encoding/base64"
	"fmt"
	"github.com/stretchr/testify/require"
	"net/http"
	"os"
	"testing"
)

func TestSonatype_GetComponentReports(t *testing.T) {
	token := os.Getenv("OSSINDEX_API_BASIC_AUTH")
	sonatype := Sonatype{
		APIBaseURL: "https://ossindex.sonatype.org/api/v3",
		APIBasicAuth: base64.StdEncoding.
			EncodeToString([]byte(
				fmt.Sprintf("%s:%s", "jiajue.huang@live.com", token))),
		HttpClient: http.Client{},
	}

	reports, err := sonatype.GetComponentReports([]string{
		"pkg:maven/log4j/log4j@1.2.17",
		"pkg:maven/log4j/log4j@2.14.0",
	})

	require.Nilf(t, err, "failed to get reports: %v", err)
	require.Equal(t, 2, len(reports))
}
