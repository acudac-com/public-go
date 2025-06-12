package blob

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
	// Reads a blob
	Read(ctx context.Context, key string) ([]byte, error)
	// Writes a blob
	Write(ctx context.Context, key string, data []byte) error
	// Writes a blob if the key does not contain any data yet
	WriteIfMissing(ctx context.Context, key string, data []byte) error
	// Removes a blob if it exists
	Remove(ctx context.Context, key string) error
	// Removes a folder and all children blobs
	RemoveFolder(ctx context.Context, folder string) error

	// Returns an io readerCloser
	Reader(ctx context.Context, key string) (io.ReadCloser, error)
	// Returns an io writerCloser
	Writer(ctx context.Context, key string) (io.WriteCloser, error)
}

type AlreadyExistsError struct {
	key string
}

func (e *AlreadyExistsError) Error() string {
	return f("blob already exists: %s", e.key)
}

type NotFoundError struct {
	key string
}

func (e *NotFoundError) Error() string {
	return f("blob not found: %s", e.key)
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

// Reads a blob from the local file system.
func (l *Fs) Read(ctx context.Context, key string) ([]byte, error) {
	path := filepath.Join(l.basePath, key)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, &NotFoundError{key}
		}
		return nil, e("reading file: %w", err)
	}
	return data, nil
}

// Writes a blob to the local file system.
func (l *Fs) Write(ctx context.Context, key string, data []byte) error {
	path := filepath.Join(l.basePath, key)
	dir := filepath.Dir(path) // Ensure directory exists
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return e("creating directory: %w", err)
	}
	return os.WriteFile(path, data, 0o644)
}

// Writes a blob to the local file system if the key does not contain any data yet
// Returns blob.AlreadyExistsError if already exists.
func (l *Fs) WriteIfMissing(ctx context.Context, key string, data []byte) error {
	path := filepath.Join(l.basePath, key)
	dir := filepath.Dir(path) // Ensure directory exists
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return e("creating directory: %w", err)
	}
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		if os.IsExist(err) {
			return &AlreadyExistsError{key}
		}
		return e("opening file with O_EXCL: %w", err)
	}
	defer f.Close()

	// If we reached here, the file was just created exclusively.
	// Now we can safely write to it.
	_, err = f.WriteAt(data, 0)
	if err != nil {
		return e("writing data: %w", err)
	}
	return nil
}

// Removes a blob from the local file system.
func (l *Fs) Remove(ctx context.Context, key string) error {
	path := filepath.Join(l.basePath, key)
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return &NotFoundError{key}
		}
		return e("removing file: %w", err)
	}
	return nil
}

// Removes a folder
func (l *Fs) RemoveFolder(ctx context.Context, folder string) error {
	path := filepath.Join(l.basePath, folder)
	if err := os.RemoveAll(path); err != nil {
		if os.IsNotExist(err) {
			return &NotFoundError{folder}
		}
		return e("removing folder: %w", err)
	}
	return nil
}

// Returns an io readerCloser for the blob at the given key.
func (l *Fs) Reader(ctx context.Context, key string) (io.ReadCloser, error) {
	path := filepath.Join(l.basePath, key)
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, &NotFoundError{key}
		}
		return nil, e("opening file: %w", err)
	}
	return file, nil
}

// Returns an io writerCloser for the blob at the given key.
func (l *Fs) Writer(ctx context.Context, key string) (io.WriteCloser, error) {
	path := filepath.Join(l.basePath, key)
	dir := filepath.Dir(path) // Ensure directory exists
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, e("creating directory: %w", err)
	}
	file, err := os.Create(path)
	if err != nil {
		return nil, e("creating file: %w", err)
	}
	return file, nil
}

// Gcs implements Storage for Google Cloud Storage.
type Gcs struct {
	bucket *storage.BucketHandle
	prefix string
}

// Returns a new Gcs blob storage instance.
func NewGcsStorage(ctx context.Context, bucket string, prefix string) (*Gcs, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, e("creating client: %w", err)
	}
	return &Gcs{client.Bucket(bucket), prefix}, nil
}

// Reads a blob from Google Cloud Storage.
func (g *Gcs) Read(ctx context.Context, key string) ([]byte, error) {
	key = path.Join(g.prefix, key)
	rc, err := g.bucket.Object(key).NewReader(ctx)
	if err != nil {
		if err.Error() == "storage: object doesn't exist" {
			return nil, &NotFoundError{key}
		}
		if apiErr, ok := err.(*googleapi.Error); ok && apiErr.Code == 404 {
			return nil, &NotFoundError{key}
		}
		return nil, e("creating reader: %w", err)
	}
	defer rc.Close()
	data, err := io.ReadAll(rc)
	if err != nil {
		if apiErr, ok := err.(*googleapi.Error); ok && apiErr.Code == 404 {
			return nil, &NotFoundError{key}
		}
	}
	return data, nil
}

// Writes a blob to Google Cloud Storage.
func (g *Gcs) Write(ctx context.Context, key string, data []byte) error {
	key = path.Join(g.prefix, key)
	wc := g.bucket.Object(key).NewWriter(ctx)

	if _, err := wc.Write(data); err != nil {
		return e("writing: %w", err)
	}
	if err := wc.Close(); err != nil {
		return e("closing writer: %w", err)
	}
	return nil
}

// Writes a blob to Google Cloud Storage if the key does not contain any data yet.
// Returns blob.AlreadyExistsError if already exists.
func (g *Gcs) WriteIfMissing(ctx context.Context, key string, data []byte) error {
	key = path.Join(g.prefix, key)
	wc := g.bucket.Object(key).If(storage.Conditions{DoesNotExist: true}).NewWriter(ctx)

	if _, err := wc.Write(data); err != nil {
		return e("writing: %w", err)
	}
	if err := wc.Close(); err != nil {
		if apiErr, ok := err.(*googleapi.Error); ok && apiErr.Code == 412 {
			return &AlreadyExistsError{key}
		}
		return e("closing writer: %w", err)
	}
	return nil
}

// Remove removes a blob from Google Cloud Storage.
func (g *Gcs) Remove(ctx context.Context, key string) error {
	key = path.Join(g.prefix, key)
	err := g.bucket.Object(key).Delete(ctx)
	if err != nil {
		if apiErr, ok := err.(*googleapi.Error); ok && apiErr.Code == 404 {
			return &NotFoundError{key}
		}
		return e("deleting object: %w", err)
	}
	return nil
}

// Removes all objects at the specified folder (prefix)
func (g *Gcs) RemoveFolder(ctx context.Context, folder string) error {
	folder = path.Join(g.prefix, folder)
	it := g.bucket.Objects(ctx, &storage.Query{Prefix: folder + "/"})
	errG, ctx := errgroup.WithContext(ctx)
	for {
		objAttrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return e("iterating objects: %w", err)
		}
		errG.Go(func() error {
			err = g.bucket.Object(objAttrs.Name).Delete(ctx)
			if err != nil {
				if apiErr, ok := err.(*googleapi.Error); ok && apiErr.Code == 404 {
					return nil // Object already deleted
				}
				return e("deleting object: %w", err)
			}
			return nil
		})
	}
	if err := errG.Wait(); err != nil {
		return e("waiting for delete operations: %w", err)
	}
	return nil
}

// Returns an io readerCloser for the blob at the given key.
func (g *Gcs) Reader(ctx context.Context, key string) (io.ReadCloser, error) {
	key = path.Join(g.prefix, key)
	rc, err := g.bucket.Object(key).NewReader(ctx)
	if err != nil {
		return nil, e("creating reader: %w", err)
	}
	return rc, nil
}

// Returns an io writerCloser for the blob at the given key.
func (g *Gcs) Writer(ctx context.Context, key string) (io.WriteCloser, error) {
	key = path.Join(g.prefix, key)
	wc := g.bucket.Object(key).NewWriter(ctx)
	if wc == nil {
		return nil, e("creating writer for key %s", key)
	}
	return wc, nil
}

// Ensure that our types satisfy the interface
var (
	_ Storage = &Fs{}
	_ Storage = &Gcs{}
)

var (
	f = fmt.Sprintf
	e = fmt.Errorf
)
