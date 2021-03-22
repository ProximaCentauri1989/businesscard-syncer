package handlers

import (
	"context"
	"log"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/radovskyb/watcher"
)

type s3Syncer struct {
	root       string
	bucketName string
	uploader   *s3manager.Uploader
}

type SyncFolderIterator struct {
	bucket    string
	fileInfos []fileInfo
	err       error
}

type fileInfo struct {
	key      string
	fullpath string
}

func (iter *SyncFolderIterator) Next() bool {
	return len(iter.fileInfos) > 0
}

// Err returns any error when os.Open is called.
func (iter *SyncFolderIterator) Err() error {
	return iter.err
}

func (iter *SyncFolderIterator) UploadObject() s3manager.BatchUploadObject {
	fi := iter.fileInfos[0]
	iter.fileInfos = iter.fileInfos[1:]
	body, err := os.Open(fi.fullpath)
	if err != nil {
		iter.err = err
	}

	extension := filepath.Ext(fi.key)
	mimeType := mime.TypeByExtension(extension)

	if mimeType == "" {
		mimeType = "binary/octet-stream"
	}

	input := s3manager.UploadInput{
		Bucket:      &iter.bucket,
		Key:         &fi.key,
		Body:        body,
		ContentType: &mimeType,
	}

	return s3manager.BatchUploadObject{
		Object: &input,
	}
}

func NewS3Syncer(region, bucket, root string) *s3Syncer {
	sess := session.New(&aws.Config{
		Region: &region,
	})
	uploader := s3manager.NewUploader(sess)
	return &s3Syncer{
		uploader: uploader,
	}
}

func (s *s3Syncer) Handle(ctx context.Context, event watcher.Event, wg *sync.WaitGroup) error {
	var err error = nil
	select {
	case <-ctx.Done():
		log.Println("Cancelation signal received. Syncing session will be skipped")
	default:
		err = s.sync()
	}
	wg.Done()
	return err
}

func (s *s3Syncer) newSyncIterator() *SyncFolderIterator {
	metadata := []fileInfo{}
	filepath.Walk(s.root, func(p string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			key := strings.TrimPrefix(p, s.root)
			metadata = append(metadata, fileInfo{key, p})
		}

		return nil
	})

	return &SyncFolderIterator{
		s.bucketName,
		metadata,
		nil,
	}
}

func (s *s3Syncer) sync() error {
	iter := s.newSyncIterator()
	if err := s.uploader.UploadWithIterator(aws.BackgroundContext(), iter); err != nil {
		return err
	}

	if err := iter.Err(); err != nil {
		return err
	}
	log.Printf("Folder '%s' succesfully synced with bucker '%s'", s.root, s.bucketName)
	return nil
}
