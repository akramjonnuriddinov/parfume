// Command push-uploads copies the local uploads folder into the production
// bucket, keeping the same paths, so the /uploads/... image links stored in
// the database keep working after the move.
//
//	AWS_BUCKET=... AWS_ENDPOINT_URL=... AWS_ACCESS_KEY_ID=... AWS_SECRET_ACCESS_KEY=... \
//	    go run ./cmd/push-uploads -dir ./uploads
package main

import (
	"context"
	"flag"
	"io/fs"
	"log"
	"mime"
	"os"
	"path/filepath"

	"parfume/backend/internal/config"
	"parfume/backend/internal/storage"
)

func main() {
	dir := flag.String("dir", "./uploads", "local uploads folder")
	flag.Parse()

	cfg := config.Load()
	if cfg.S3.Bucket == "" {
		log.Fatal("AWS_BUCKET is not set")
	}
	bucket, err := storage.New(cfg.S3, "")
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	var count int
	err = filepath.WalkDir(*dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || d.Name() == ".gitkeep" {
			return err
		}
		rel, _ := filepath.Rel(*dir, path)
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		st, _ := f.Stat()
		if err := bucket.Put(ctx, filepath.ToSlash(rel), f, st.Size(), mime.TypeByExtension(filepath.Ext(path))); err != nil {
			return err
		}
		count++
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("uploaded %d files", count)
}
