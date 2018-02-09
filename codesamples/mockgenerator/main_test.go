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
	"github.com/stretchr/testify/mock"
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
	} {
		// Set up and call GetMessage
		buf := &bytes.Buffer{}
		getter := New(log.New(buf, "test", log.LstdFlags), test.useRemoteFS, test.isCanary, "accesskey", "accesssecret")
		fr := &fakeRemote{data: test.data, err: test.remoteErr}
		fr.On("GetRemoteMessage", "accesskey", "accesssecret", "host", "bucket", "key").Return(nil, nil, fmt.Errorf("an error"))
		getter.RemoteGetter = fr
		getter.LocalGetter = &fakeLocal{data: test.data, err: test.localErr}
		fh, source, err := getter.GetMessage("localpath", test.host, test.bucket, test.key)

		// make sure that everything was as expected
		assert.Equal(t, test.expectedSource, source, fmt.Sprintf("test %q", test.name))
		assert.Equal(t, test.expectedErr, err, fmt.Sprintf("test %q", test.name))

		fr.AssertCalled(t, "GetRemoteMessage", "accessKey", "accessSecret", "host", "bucket", "key")

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
	mock.Mock
	data []byte
	err  error
}

func (f *fakeRemote) GetRemoteMessage(accessKey, accessSecret, host, bucket, key string) (io.ReadCloser, error) {
	f.Called(accessKey, accessSecret, host, bucket, key)
	return ioutil.NopCloser(bytes.NewReader(f.data)), f.err
}

type fakeLocal struct {
	mock.Mock
	data []byte
	err  error
}

func (f *fakeLocal) Open(localPath string) (io.ReadCloser, error) {
	return ioutil.NopCloser(bytes.NewReader(f.data)), f.err
}
