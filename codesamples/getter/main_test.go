package getter

// TestGetMessage verifies that we can get remote and local files and that we handle errors
import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetMessage(t *testing.T) {
	for _, test := range []struct {
		// name gives us a test indentifier for if a test fails, we can know which one
		name string
		// based on config for if this system should allow for remote file system access
		useRemoteFS bool
		// based on configured list of "canary" servers which should access the remote file system
		isCanary bool
		// represents the data in the file
		data []byte
		// the GetMessage method will report back if it was remote or local for the source of the file
		expectedSource string
		// set what the remote err should be
		remoteErr error
		// set what the local err should be
		localErr error
		// the error returned from GetMessage depending on how remoteErr and localErr behaved
		expectedErr error
		// allow us to inspect any error logs generated
		expectedLogs []string
		// parameters into GetMessage. We expect errors if any are blank
		bucket string
		key    string
		host   string
	}{
		{
			name:           "should use remote fs",
			useRemoteFS:    true,
			isCanary:       true,
			data:           []byte("file data"),
			bucket:         "bucket",
			key:            "key",
			host:           "host",
			expectedSource: SourceRemote,
			remoteErr:      nil,
			localErr:       nil,
			expectedErr:    nil,
			expectedLogs:   []string{},
		},
		{
			name:           "should use local - config dont use remote file system",
			useRemoteFS:    false,
			isCanary:       true,
			data:           []byte("file data"),
			bucket:         "bucket",
			key:            "key",
			host:           "host",
			expectedSource: SourceLocal,
			remoteErr:      nil,
			localErr:       nil,
			expectedErr:    nil,
			expectedLogs:   []string{},
		},
		{
			name:           "should use local - config not a canary",
			useRemoteFS:    true,
			isCanary:       false,
			data:           []byte("file data"),
			bucket:         "bucket",
			key:            "key",
			host:           "host",
			expectedSource: SourceLocal,
			remoteErr:      nil,
			localErr:       nil,
			expectedErr:    nil,
			expectedLogs:   []string{},
		},
		{
			name:           "should use local - config not a canary + config don't use remote files ystem",
			useRemoteFS:    false,
			isCanary:       false,
			data:           []byte("file data"),
			bucket:         "bucket",
			key:            "key",
			host:           "host",
			expectedSource: SourceLocal,
			remoteErr:      nil,
			localErr:       nil,
			expectedErr:    nil,
			expectedLogs:   []string{},
		},
		{
			name:           "error remote - fall back to local",
			useRemoteFS:    true,
			isCanary:       true,
			data:           []byte("file data"),
			bucket:         "bucket",
			key:            "key",
			host:           "host",
			expectedSource: SourceLocal,
			remoteErr:      fmt.Errorf("unable to remote"),
			localErr:       nil,
			expectedErr:    nil,
			expectedLogs:   []string{"falling back to local source"},
		},
		{
			name:           "error remote and local - report back local error and log remote failure",
			useRemoteFS:    true,
			isCanary:       true,
			data:           []byte("file data"),
			bucket:         "bucket",
			key:            "key",
			host:           "host",
			expectedSource: SourceLocal,
			remoteErr:      fmt.Errorf("falling back to local source"),
			localErr:       fmt.Errorf("unable to read from disk"),
			expectedErr:    fmt.Errorf("unable to read from disk"),
			expectedLogs:   []string{"falling back to local source"},
		},
		{
			name:           "error remote for bad bucket, use local",
			useRemoteFS:    true,
			isCanary:       true,
			data:           []byte("file data"),
			bucket:         "",
			key:            "key",
			host:           "host",
			expectedSource: SourceLocal,
			remoteErr:      nil,
			localErr:       nil,
			expectedErr:    nil,
			expectedLogs:   []string{"falling back to local source - missing fields"},
		},
		{
			name:           "error remote for bad key, use local",
			useRemoteFS:    true,
			isCanary:       true,
			data:           []byte("file data"),
			bucket:         "bucket",
			key:            "",
			host:           "host",
			expectedSource: SourceLocal,
			remoteErr:      nil,
			localErr:       nil,
			expectedErr:    nil,
			expectedLogs:   []string{"falling back to local source - missing fields"},
		},
		{
			name:           "error remote for bad host, use local",
			useRemoteFS:    true,
			isCanary:       true,
			data:           []byte("file data"),
			bucket:         "bucket",
			key:            "key",
			host:           "",
			expectedSource: SourceLocal,
			remoteErr:      nil,
			localErr:       nil,
			expectedErr:    nil,
			expectedLogs:   []string{"falling back to local source - missing fields"},
		},
	} {
		// Set up and call GetMessage
		buf := &bytes.Buffer{}
		getter := New(log.New(buf, "test", log.LstdFlags), test.useRemoteFS, test.isCanary, "accesskey", "accesssecret")
		getter.RemoteGetter = &fakeRemote{data: test.data, err: test.remoteErr}
		getter.LocalGetter = &fakeLocal{data: test.data, err: test.localErr}
		fh, source, err := getter.GetMessage("localpath", test.host, test.bucket, test.key)

		// make sure that everything was as expected
		assert.Equal(t, test.expectedSource, source, fmt.Sprintf("test %q", test.name))
		assert.Equal(t, test.expectedErr, err, fmt.Sprintf("test %q", test.name))

		for _, expected := range test.expectedLogs {
			assert.Contains(t, buf.String(), expected, fmt.Sprintf("test %q", test.name))
		}

		if err != nil {
			// no need to verify fh contents, go to next test case
			continue
		}

		// verify message content
		data, err := ioutil.ReadAll(fh)
		assert.NoError(t, err, fmt.Sprintf("test %q", test.name))
		assert.Equal(t, test.data, data, fmt.Sprintf("test %q", test.name))
		fh.Close()
	}
}

type fakeRemote struct {
	data []byte
	err  error
}

func (f *fakeRemote) GetRemoteMessage(accessKey, accessSecret, host, bucket, key string) (io.ReadCloser, error) {
	return ioutil.NopCloser(bytes.NewReader(f.data)), f.err
}

type fakeLocal struct {
	data []byte
	err  error
}

func (f *fakeLocal) Open(localPath string) (io.ReadCloser, error) {
	return ioutil.NopCloser(bytes.NewReader(f.data)), f.err
}
