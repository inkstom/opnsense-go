package captiveportal

import (
	"os"
	"testing"

	"github.com/browningluke/opnsense-go/pkg/api"
)

func newTestController(t *testing.T) *Controller {
	t.Helper()
	return &Controller{Api: api.NewClient(api.Options{
		Uri:           os.Getenv("OPNSENSE_URI"),
		APIKey:        os.Getenv("OPNSENSE_API_KEY"),
		APISecret:     os.Getenv("OPNSENSE_API_SECRET"),
		AllowInsecure: true,
		MaxBackoff:    30,
		MinBackoff:    1,
		MaxRetries:    4,
	})}
}
