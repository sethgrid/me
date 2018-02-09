package getter

import (
	"bytes"
	"log"
	"testing"
)

func TestEmbeddedMock(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := log.New(buf, "testing", log.LstdFlags)
	messageGetter := New(logger, true, true, "key", "secret")
	fh, source, err := messageGetter.GetMessage("localpath", "host", "bucket", "key")

	messageGetter.AssertCalled(t, "GetMesssage", "host", "key", "secret", false)
	messageGetter.AssertNumberOfCalls(t, "New", 1)

	// go on to verify responses. Assigning to blank identifier to avoid "declared and not used" compilation errors
	_ = fh
	_ = source
	_ = err
}
