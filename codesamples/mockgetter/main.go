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

// Getter contains unexported fields allowing the local or remote fetching of files
type Getter struct {
	logger           *log.Logger
	useRemoteFS      bool
	isRemoteFSCanary bool
	accessKey        string
	accessSecret     string
}

// New creates a instatialized Getter that can get files locally or remotely.
// useRemoteFS tells us if the service is configured to use the remote file system.
// isRemoteFSCanaryHost says that this particular host in this cluster should be attempting remote file system work.
// accessKey and accessSecret are authentication parts for the remote file system.
func New(l *log.Logger, useRemoteFS, isRemoteFSCanaryHost bool, accessKey, accessSecret string) *Getter {
	return &Getter{
		logger:           l,
		useRemoteFS:      useRemoteFS,
		isRemoteFSCanary: isRemoteFSCanaryHost,
		accessKey:        accessKey,
		accessSecret:     accessSecret,
	}
}

// GetMessage will reach out to s3 or use the local file system to retrieve a file / message
func (g *Getter) GetMessage(localPath, host, bucket, key string) (io.ReadCloser, string, error) {
	if g.useRemoteFS && g.isRemoteFSCanary && host != "" && key != "" && bucket != "" {
		// we have everything we need to do remote fs stuff
		var err error
		var client *minio.Client
		var obj *minio.Object

		client, err = minio.NewV2(host, g.accessKey, g.accessSecret, false)
		if err != nil {
			err = errors.Wrap(err, "unable to get remote fs client")
		} else {
			obj, err = client.GetObject(bucket, key)
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

	} else if g.useRemoteFS && g.isRemoteFSCanary {
		// we want to do remote fs stuff, but host, bucket, or key are messed up
		g.logger.Println("falling back to local source - missing fields")
	}

	fh, err := os.Open(localPath)
	if err != nil {
		return nil, "", err
	}

	return fh, SourceLocal, nil
}
