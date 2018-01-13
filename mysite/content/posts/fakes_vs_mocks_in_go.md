+++
date = "2018-01-01T18:02:55-07:00"
draft = false
title = "Fakes vs Mock in Go"

+++

I really enjoy writing tests for my code, especially unit tests. The sense of confidence it gives me is great. Picking up something I've not worked on in a long time and being able to run the unit and integration tests gives me the knowledge that I can ruthlessly refactor if needed and that, as long as my tests continue to pass, I will still have functional software afterword. Unit tests guide code design and allows us to quickly test that errors and logic flows work as intended. With that, I want to posit something perhaps a bit more contriversial: when writing unit tests, don't use mocks.

Let's get some definitions on the table. What do I mean by mocks and what should you use instead? This post is focused on working in Go, and so my slant on these words are in the context of Go. I think that using the right words and sharing common definitions is important, especially for technical topics. When I say "mocks," I am specifically referring to the term "Mock Object" as intended in the paper [Endo-Testing: Unit Testing with Mock Objects (2000)](https://www.google.com/url?sa=t&rct=j&q=&esrc=s&source=web&cd=1&cad=rja&uact=8&ved=0ahUKEwiSmsrW6NXYAhVO4WMKHdNDCDEQFggnMAA&url=https%3A%2F%2Fwww2.ccs.neu.edu%2Fresearch%2Fdemeter%2Frelated-work%2Fextreme-programming%2FMockObjectsFinal.PDF&usg=AOvVaw2-liwRQ7dpXm4D4krnRFry]. That is to say, mock objects are where we "replace domain code with dummy implementations that both emulate real functionality and enforce assertions about the behaviour of our code." Stated a little shorter, mocks assert behavior. I advocate for "Fakes rather than Mocks." A fake is a kind of test double that may contain business behavior. To get more clarification on the differences in different test structures, check out (The Little Mocker)[https://8thlight.com/blog/uncle-bob/2014/05/14/TheLittleMocker.html). Fakes are mearly structs that fit an interface and are a form of dependency injection where we control the behavior. Let's dive in with an example that is a bit more advanced than testing a sum function. However, I need to give you some context so you can more easily understand the code.

We have an application that needs to be able read a file. This file could be on a remote file system (like S3) or it could be local. At SendGrid, we have some work where we've traditionally had files on the local file system, but for reasons around higher availability and better throughput, we are moving these files over to S3. Because this involves important data, and we will find ourselves in a position where some nodes process local files and others process remote files (with a local copy of the file as fallback) we are creating a solution that allows us to toggle back and forth as we gain confidence in the new remote file system.

With that out of the way, we have a package that needs to read files. We refer to it as `package foo`. It has a sub package, `package getter`. We need to ensure that `package getter` can get files either from the remote filesystem or the local filesystem, and fallback to local when the remote fails.

The basic (if naive) approach is that `package foo` will call `getter.New(...)` and pass it the information needed for setting up remote file getting. The returned value will then be able to call `GetMessage(...)` with the parameters needed for locating the remote or local file.

```go
type Getter struct{
	logger 				*log.Logger
	useRemoteFS			bool
	isRemoteFSCanary	bool
	accessKey 			string
	accesssSecret 		string
}

func New(l *log.Logger, useRemoteFS, isRemoteFSCanaryHost bool, accessKey, accessSecret string) *Getter {
	return &Getter{
		logger:               l,
		useRemoteFS:          useRemoteFS,
		isRemoteFSCanaryHost: isRemoteFSCanaryHost,
		accessKey:            accessKey,
		accessSecret:         accessSecret,
	}
}
```

This gives us our basic structure. When we create a the new `Getter`, we set things that are needed for any potential remote file getting (thus the access keys and secret) and we also pass in some values that originate in our application configuration: useRemoteFS and isRemoteFSCanary. The former specifies if this service has the config flag turned on to allow reading from a remote file system. The ladder specifies if this specific node should be communicating with that remote file system.

We now need to give it some basic functionality. Note, this is a non-finished example. Please excuse the nested if-statements. They will go away.

```go
const (
	SourceLocal  = "local"
	SourceRemote = "remote"
)

func GetMessage(localPath, host, bucket, key string) (io.ReadCloser, string, error) {
	if g.useRemoteFS && g.isRemoteFSCanaryHost && host != "" && key != "" && bucket != "" {
		// we have everything we need to do remote fs stuff
		var err error
		var client *minio.Client
		var obj *minio.Object

		client, err = minio.NewV2(host, accessKey, accessSecret, false)
		if err != nil {
			err = errors.Wrap(err, "unable to get remote fs client")
		} else {
			obj, err := client.GetObject(bucket, key)
			if err != nil {
				err = errors.Wrap(err, "unable to get remote object")
			} else {
				_, err = obj.Stat()
				if err != nil {
					err = errors.Wrap(err, "unable to get remote file info")
				} else {
					return obj, SourceRemote, nil
				}
			}
		}
		// if we get here, we are falling back to local disk

	} else if g.useRemoteFS && g.isRemoteFSCanaryHost {
		// we want to do remote fs stuff, but host, bucket, or key are messed up
		logger.Println("falling back to local source - missing fields")
	}

	fh, err := os.Open(localPath)
	if err != nil {
		return nil, "", err
	}

	return fh, SourceLocal, nil
}
```

The basic idea here is that if we are configured to read from the remote file system and we get remote file system details (host, bucket, and key), then we should attempt to read from the remote file system. In the event that that process fails or if we are not configured to do remote file system work, we should fall back to local disk, where a back up copy of the file would be residing. Remember, this fall back design is because we are testing out the new remote file system. After we have confidence in it, we will shift all file reading out to the remote file system and remove references to reading from the local file system.

You may have noticed that this code is not as testable as it should be. Note that to verify how it works, we actually need to hit not only the local file system, but the remote file system too. If we were doing an integration test only and set up some docker-fu to have an s3 instance, then we could just do with the integration tests. However, I believe that unit tests help us design more robust software and easily test alternate code paths. We should save integration tests for larger "does it really work" kinds of tests. Don't worry, we will talk about an integration test for this too, but let's focus on the unit tests for now.

How can we make this code more unit testable? There are two schools of thought. One is to use a mock generator and generate the filesystem calls and the minio client calls. It turns out that mocking the minio client is not straight forward because you have a typed client that returns a typed object. These concrete types are not interfaces so using generators is out of the question. We can make our own interfaces and wrap the minio actions, but we will not be able to override the minio client. While we could rip up the interfaces into a strange russian doll of MinioClienter to MinioObjector, and MinioStater, I say that there is a better way. If we restructure our code to be more testable, we don't need additional imports and cruft and there will be no need for knowing additional testing DSLs to confidently test the interfaces. Testing code will just be normal Go code. Let's do it.

What is it that we need to test? We need to test that we can get files remotely or locally and some failure paths, such as remote failures having us fall back to local file sources. Let's refactor that initial approach and pull out the minio client into a remote getter. While we are doing that, let's do the same to our code for local file reading, and make a local getter.

```go
type RemoteGetter interface {
	GetRemoteMessage(accessKey, accessSecret, host, bucket, key string) (io.ReadCloser, error)
}

type minioWrapper struct{}

func (_ *minioWrapper) GetRemoteMessage(accessKey, accessSecret, host, bucket, key string) (io.ReadCloser, error) {
	client, err := minio.NewV2(host, accessKey, accessSecret, false)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get remote fs client")
	}

	obj, err := client.GetObject(bucket, key)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get remote object")
	}
	_, err = obj.Stat()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get remote file info")
	}

	return obj, nil
}

type LocalGetter interface {
	Open(localPath string) (io.ReadCloser, error)
}

type osFile struct{}

func (f *osFile) Open(localPath string) (io.ReadCloser, error) {
	return os.Open(localPath)
}
```

With these abstractions in place, we can refactor our inital implementation. We are going to put the `localGetter` and `remoteGetter` onto the `Getter` struct and refactor `GetMessage` to use them. Here is everything put together and refactored:

```go
package getter

import (
	"io"
	"os"
	"log"

	"github.com/minio/minio-go"
	"github.com/pkg/errors"
)

const (
	// SourceLocal signifies if a file or error came from local disk
	SourceLocal  = "local"
	// SourceRemote signifies if a file or error came from the remote disk
	SourceRemote = "remote"
)

// MessageGetter allows us to get a ReadCloser, the source (remote/local), or an error when attempting to get
// a message from either local or remote storage
type MessageGetter interface {
	GetMessage(localPath, host, bucket, key string) (io.ReadCloser, string, error)
}

// Getter contains unexported fields allowing the local or remote fetching of files
type Getter struct {
	logger               *log.Logger
	useRemoteFS          bool
	isRemoteFSCanaryHost bool
	accessKey            string
	accessSecret         string

	RemoteGetter RemoteGetter
	LocalGetter  LocalGetter
}

// New creates a instatialized Getter that can get files locally or remotely.
// useRemoteFS tells us if the service is configured to use the remote file system.
// isRemoteFSCanaryHost says that this particular host in this cluster should be attempting remote file system work.
// accessKey and accessSecret are authentication parts for the remote file system.
func New(l *log.Logger, useRemoteFS, isRemoteFSCanaryHost bool, accessKey, accessSecret string) *Getter {
	return &Getter{
		logger:               l,
		useRemoteFS:          useRemoteFS,
		isRemoteFSCanaryHost: isRemoteFSCanaryHost,
		accessKey:            accessKey,
		accessSecret:         accessSecret,
		RemoteGetter:         &minioWrapper{},
		LocalGetter:          &osFile{},
	}
}

// GetMessage will reach out to s3 / rados gateway or use the local file system to retrieve an email message
func (g *Getter) GetMessage(localPath, host, bucket, key string) (io.ReadCloser, string, error) {
	if g.useRemoteFS && g.isRemoteFSCanaryHost && host != "" && key != "" && bucket != "" {
		// we have everything we need to do remote fs stuff
		fh, err := g.RemoteGetter.GetRemoteMessage(g.accessKey, g.accessSecret, host, bucket, key)
		if err == nil {
			return fh, SourceRemote, nil
		}

		logger.Println("falling back to local source - %v", err)
	} else if g.useRemoteFS && g.isRemoteFSCanaryHost {
		// we want to do remote fs stuff, but host, bucket, or key are messed up
		logger.Println("falling back to local source - missing fields")
	}

	fh, err := g.LocalGetter.Open(localPath)
	if err != nil {
		return nil, SourceLocal, err
	}

	return fh, SourceLocal, nil
}

type RemoteGetter interface {
	GetRemoteMessage(accessKey, accessSecret, host, bucket, key string) (io.ReadCloser, error)
}

type minioWrapper struct{}
func (_ *minioWrapper) GetRemoteMessage(accessKey, accessSecret, host, bucket, key string) (io.ReadCloser, error) {
	client, err := minio.NewV2(host, accessKey, accessSecret, false)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get remote fs client")
	}

	obj, err := client.GetObject(bucket, key)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get remote object")
	}
	_, err = obj.Stat()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get remote file info")
	}

	return obj, nil
}

type LocalGetter interface {
	Open(localPath string) (io.ReadCloser, error)
}

type osFile struct{}
func (f *osFile) Open(localPath string) (io.ReadCloser, error) {
	return os.Open(localPath)
}

```

This new, refactored code is much more testable. Because we take interfaces as parameters on the `Getter` struct, we can change out the concrete types. Instead of mocking OS calls or needing a full mocking of the minio client, we just need two fakes. A `fakeLocalGetter` and a `fakeRemoteGetter`. These fakes have some properties on them that let us specify what they return. We will be able to return the file data or any error we like from these fakes and we can verify that the calling `GetMessage` method handles the data and errors as we intended.

With this in mind, the heart of the tests become:
```go
buf := &bytes.Buffer{}
// some parameters to test:
configSystemShouldUseRemoteFileSystem := true
configThisHostShouldUseRemoteFileSystem := true

// create a new Getter with test params
getter := New(log.New(buf, "test", log.LstdFlags), configSystemShouldUseRemoteFileSystem, configThisHostShouldUseRemoteFileSystem, "accesskey", "accesssecret")

// override the remote and local getters to use fakes that return some error
getter.RemoteGetter = &fakeRemote{data: []byte("file data"), err: fmt.Errorf("some error on the remote file system")}
getter.LocalGetter = &fakeLocal{data: []byte("file data"), err: fmt.Errorf("some error on the local file system")}
fh, source, err := getter.GetMessage("localpath", "host", "bucket", "key")

// assert against the return values that include the file data, the source of the file data (or error), and the error
```

With this basic structure, we can wrap it all up in table driven tests. Each case in the table of tests will either be testing for local or remote file access. We will be able to inject an error at either remote or local file access. We can verify propagated errors, that the file contents are passed up, and that expected log entries are present. I went ahead and included all potential test cases and permutations in the one table driven test. These could have been broken up to make each test case a tad less verbose, but helps us adhere to DRY.

The full test:
```go
// TestGetMessage verifies that we can get remote and local files and that we handle errors
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
			expectedLogs:   []string{"falling back to local source - missing fields", `"rg_bucket":""`, `"rg_key":"key"`, `"rg_host":"host"`},
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
			expectedLogs:   []string{"falling back to local source - missing fields", `"rg_bucket":"bucket"`, `"rg_key":""`, `"rg_host":"host"`},
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
			expectedLogs:   []string{"falling back to local source - missing fields", `"rg_bucket":"bucket"`, `"rg_key":"key"`, `"rg_host":""`},
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
			assert.Contains(t, logBuf.String(), expected, fmt.Sprintf("test %q", test.name))
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
```

Nifty, eh? We have full control of how we want GetMessage to behave, and we can assert against the results. We've designed our code to be unit-test friendly and can now verify success and error paths implemented in the `GetMessage` method. We did this by writing plain ol' Go code that any developer familiar with Go should be able to understand and extend when needed.

But what about mocks? What would the mocks buy us that we don't get here? A great question could be, "how do you know you called the s3 client with the correct parameters? With mocks, I can ensure that I passed the key value to the key parameter, and not the bucket parameter." That is a valid concern and it should be covered under a test _somewhere_. The testing approach that I advocate here does not verify that you called the minio client with the bucket and key parameters in the right order.

A great quote I recently read stated, 'Mocking introduces assumptions, which introduce risk' (https://charemza.name/blog/posts/methodologies/testing/questions-to-ask-yourself-when-writing-tests/). You are assuming the client library is implemented right, you are assuming all boundaries are solid. Mocking it out only mocks assumptions and likely makes your tests more brittle and subject to change when you update the code. When the rubber meets the road, we are going to have to verify that we are actually using the minio client correctly. There is no need for a unit test to cover the exact implementation. That is what integration tests are for. The unit test has guided our code design and allows us to quickly test that errors and logic flows work as designed. To ensure proper use of the client, you have to actually call the client. These tests might be in docker or in a staging environment. They might operate of the built binary or not. But the test of the integration with the remote file system will need to happen somewhere prior to produciton.

For some, they feel that this is not enough unit test coverage. They will insist on either russian doll style interfaces, maybe like the following:
```go

type s3ClientMaker interface {
	NewV2(string, string, string bool) objectGetter
}

type s3ObjectGetter interface {
	GetObject(string, string) (stater, error)
}

type s3Stater interface {
	Stat() (interface{}, error)
}
```

And then they might pull out each part of the minio client into each wrapper and then use a mock generator (because adding dependencies to builds and tests, why not?), and at the end, we will be able to say something like, `myClientMock.ExpectsCall("GetObject").Returns(mockObject).NumberOfCalls(1).WithArgs(key, bucket)` (if you can recall the correct incantation for this specific DSL). This would be a lot of extra abstraction tied directly to the implementation choice of using the minio client causing brittle tests for when we find out we need to change out clients for whatever reason in the future. This adds to end-to-end code development time, adds to code complexity and reduces readability, potentially increases dependencies on mock generators, and gives us the dubious additional value of knowing if we mixed up the bucket and key parameters of which we will discover in integration testing. Further, as more and more objects get introduced, the coupling gets tighter and tighter. We might have made a logger mock and later we start having a metrics mock. Before you know it, you are adding a log entry or a new metric and you just broke umpteen tests that did not expect an additional metric to come through. The last time I was bit by this in Go, the mocking framework would not even tell me what test or file was failing as it paniced and died a horrible death because it came across a new metric (this required binary searching the tests by commenting them out to be able to find where we needed to alter the mock behavior).

We've shown that we can guide design and ensure proper code and error paths are followed with simple use of interfaces in Go. By writing simple fakes that adhere to the interfaces, we can see that we do not need mocks, mocking frameworks, or mock generators to create code designed for testing. We've also noted that unit testing is not everything, and you must write integration tests to ensure that systems are properly integrated with one another.
