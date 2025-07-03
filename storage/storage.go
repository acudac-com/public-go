package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"

	"cloud.google.com/go/storage"
	"golang.org/x/sync/errgroup"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/iterator"
)

// A simplified interface for interacting with blob storage.
type Storage interface {
	// Reads an object and returns whether the object was found.
	Read(ctx context.Context, key string) ([]byte, bool)
	// Writes an object
	Write(ctx context.Context, key string, data []byte)
	// Writes an object if the key does not contain any data yet. Returns true if
	// the data was written.
	WriteIfMissing(ctx context.Context, key string, data []byte) bool
	// Removes an object if it exists and return whether any object was removed.
	Remove(ctx context.Context, key string) bool
	// Removes a folder and all children objects and returns the number of
	// removed objects
	RemoveFolder(ctx context.Context, folder string)

	// Returns an io readerCloser if the object exists.
	Reader(ctx context.Context, key string) (io.ReadCloser, bool)
	// Returns an io writerCloser
	Writer(ctx context.Context, key string) io.WriteCloser
}

// Returned as reader if object's do not exist
type EmptyReaderCloser struct{}

func (e *EmptyReaderCloser) Read(p []byte) (n int, err error) {
	return 0, io.EOF
}

func (e *EmptyReaderCloser) Close() error {
	return nil
}

// Implements the Storage interface for the local file system.
type Fs struct {
	basePath string // Base path where blobs will be stored.
}

// Returns a new Fs instance.
func NewFsStorage(basePath string) *Fs {
	return &Fs{
		basePath: basePath,
	}
}

func (l *Fs) Read(ctx context.Context, key string) ([]byte, bool) {
	path := filepath.Join(l.basePath, key)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false
		}
		panic(fmt.Errorf("reading file (%s): %w", path, err))
	}
	return data, true
}

func (l *Fs) Write(ctx context.Context, key string, data []byte) {
	path := filepath.Join(l.basePath, key)
	dir := filepath.Dir(path) // Ensure directory exists
	if err := os.MkdirAll(dir, 0o755); err != nil {
		panic(fmt.Errorf("creating directory (%s): %w", dir, err))
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		panic(fmt.Errorf("writing object (%s): %w", path, err))
	}
}

func (l *Fs) WriteIfMissing(ctx context.Context, key string, data []byte) bool {
	path := filepath.Join(l.basePath, key)
	dir := filepath.Dir(path) // Ensure directory exists
	if err := os.MkdirAll(dir, 0o755); err != nil {
		panic(fmt.Errorf("creating directory (%s): %w", dir, err))
	}
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		if os.IsExist(err) {
			return false
		}
		panic(fmt.Errorf("opening file (%s) with O_EXCL: %w", path, err))
	}
	defer f.Close()

	_, err = f.WriteAt(data, 0)
	if err != nil {
		panic(fmt.Errorf("writing to file (%s): %w", path, err))
	}
	return true
}

func (l *Fs) Remove(ctx context.Context, key string) bool {
	path := filepath.Join(l.basePath, key)
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return false
		}
		panic(fmt.Errorf("removing file (%s): %w", path, err))
	}
	return true
}

func (l *Fs) RemoveFolder(ctx context.Context, folder string) {
	path := filepath.Join(l.basePath, folder)
	if err := os.RemoveAll(path); err != nil {
		if os.IsNotExist(err) {
			return
		}
		panic(fmt.Errorf("removing folder (%s): %w", path, err))
	}
}

// Returns an io readerCloser for the blob at the given key.
func (l *Fs) Reader(ctx context.Context, key string) (io.ReadCloser, bool) {
	path := filepath.Join(l.basePath, key)
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &EmptyReaderCloser{}, false
		}
		panic(fmt.Errorf("opening file: %w", err))
	}
	return file, true
}

// Returns an io writerCloser for the blob at the given key.
func (l *Fs) Writer(ctx context.Context, key string) io.WriteCloser {
	path := filepath.Join(l.basePath, key)
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		panic(fmt.Errorf("creating dir (%s): %w", path, err))
	}
	file, err := os.Create(path)
	if err != nil {
		panic(fmt.Errorf("creating file (%s): %w", path, err))
	}
	return file
}

// Gcs implements Storage for Google Cloud Storage.
type Gcs struct {
	bucket *storage.BucketHandle
	prefix string
}

// Returns a new Gcs blob storage instance.
func NewGcsStorage(bucket string, prefix string) *Gcs {
	client, err := storage.NewClient(context.Background())
	if err != nil {
		panic(fmt.Errorf("creating gcs client: %w", err))
	}
	return &Gcs{client.Bucket(bucket), prefix}
}

// Reads a blob from Google Cloud Storage.
func (g *Gcs) Read(ctx context.Context, key string) ([]byte, bool) {
	key = path.Join(g.prefix, key)
	rc, err := g.bucket.Object(key).NewReader(ctx)
	if err != nil {
		if err.Error() == "storage: object doesn't exist" {
			return nil, false
		}
		if apiErr, ok := err.(*googleapi.Error); ok && apiErr.Code == 404 {
			return nil, false
		}
		panic(fmt.Errorf("creating gcs reader for %s: %w", key, err))
	}
	defer rc.Close()
	data, err := io.ReadAll(rc)
	if err != nil {
		if apiErr, ok := err.(*googleapi.Error); ok && apiErr.Code == 404 {
			return nil, false
		}
	}
	return data, true
}

// Writes a blob to Google Cloud Storage.
func (g *Gcs) Write(ctx context.Context, key string, data []byte) {
	key = path.Join(g.prefix, key)
	wc := g.bucket.Object(key).NewWriter(ctx)
	if _, err := wc.Write(data); err != nil {
		panic(fmt.Errorf("writing gcs object (%s): %w", key, err))
	}
	if err := wc.Close(); err != nil {
		panic(fmt.Errorf("closing writer: %w", err))
	}
}

// Writes a blob to Google Cloud Storage if the key does not contain any data yet.
// Returns blob.AlreadyExistsError if already exists.
func (g *Gcs) WriteIfMissing(ctx context.Context, key string, data []byte) bool {
	key = path.Join(g.prefix, key)
	wc := g.bucket.Object(key).If(storage.Conditions{DoesNotExist: true}).NewWriter(ctx)

	if _, err := wc.Write(data); err != nil {
		panic(fmt.Errorf("writing gcs object (%s): %w", key, err))
	}
	if err := wc.Close(); err != nil {
		if apiErr, ok := err.(*googleapi.Error); ok && apiErr.Code == 412 {
			return false
		}
		panic(fmt.Errorf("closing gcs writer for %s: %w", key, err))
	}
	return true
}

// Remove removes a blob from Google Cloud Storage.
func (g *Gcs) Remove(ctx context.Context, key string) bool {
	key = path.Join(g.prefix, key)
	err := g.bucket.Object(key).Delete(ctx)
	if err != nil {
		if apiErr, ok := err.(*googleapi.Error); ok && apiErr.Code == 404 {
			return false
		}
		panic(fmt.Errorf("deleting gcs object (%s): %w", key, err))
	}
	return true
}

// Removes all objects at the specified folder (prefix)
func (g *Gcs) RemoveFolder(ctx context.Context, folder string) {
	folder = path.Join(g.prefix, folder)
	it := g.bucket.Objects(ctx, &storage.Query{Prefix: folder + "/"})
	errG, ctx := errgroup.WithContext(ctx)
	for {
		objAttrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			panic(fmt.Errorf("iterating gcs objects: %w", err))
		}
		errG.Go(func() error {
			err = g.bucket.Object(objAttrs.Name).Delete(ctx)
			if err != nil {
				if apiErr, ok := err.(*googleapi.Error); ok && apiErr.Code == 404 {
					return nil // Object already deleted
				}
				return fmt.Errorf("deleting object: %w", err)
			}
			return nil
		})
	}
	if err := errG.Wait(); err != nil {
		panic(fmt.Errorf("waiting for gcs delete err group: %w", err))
	}
}

// Returns an io readerCloser for the blob at the given key.
func (g *Gcs) Reader(ctx context.Context, key string) (io.ReadCloser, bool) {
	key = path.Join(g.prefix, key)
	rc, err := g.bucket.Object(key).NewReader(ctx)
	if err != nil {
		if err.Error() == "storage: object doesn't exist" {
			return &EmptyReaderCloser{}, false
		}
		if apiErr, ok := err.(*googleapi.Error); ok && apiErr.Code == 404 {
			return &EmptyReaderCloser{}, false
		}
		panic(fmt.Errorf("creating gcs reader for %s: %w", key, err))
	}
	return rc, true
}

// Returns an io writerCloser for the blob at the given key.
func (g *Gcs) Writer(ctx context.Context, key string) io.WriteCloser {
	key = path.Join(g.prefix, key)
	wc := g.bucket.Object(key).NewWriter(ctx)
	if wc == nil {
		panic(fmt.Errorf("creating gcs writer for key %s", key))
	}
	return wc
}

// Ensure that our types satisfy the interface
var (
	_ Storage = &Fs{}
	_ Storage = &Gcs{}
)
