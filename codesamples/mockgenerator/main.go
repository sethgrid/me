package getter

import (
	"io"
	"log"
	"os"

	"github.com/minio/minio-go"
	"github.com/pkg/errors"
)

const (
	// SourceLocal signifies if a file or error came from local disk
	SourceLocal = "local"
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

		g.logger.Printf("falling back to local source - %v", err)
	} else if g.useRemoteFS && g.isRemoteFSCanaryHost {
		// we want to do remote fs stuff, but host, bucket, or key are messed up
		g.logger.Println("falling back to local source - missing fields")
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

type s3ClientMaker interface {
	NewV2(string, string, string bool) objectGetter
}

type s3ObjectGetter interface {
	GetObject(string, string) (stater, error)
}

type s3Stater interface {
	Stat() (interface{}, error)
}
