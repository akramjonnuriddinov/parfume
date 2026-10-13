package api

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const maxUploadSize = 5 << 20

var allowedImageTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
	"image/gif":  ".gif",
}

func (h *Handler) upload(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize+1024)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		writeError(w, http.StatusRequestEntityTooLarge, "file is too large (max 5MB)")
		return
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "file is required")
		return
	}
	defer file.Close()

	// Detect type from content, not from the client-provided name/header.
	head := make([]byte, 512)
	n, _ := io.ReadFull(file, head)
	ext, ok := allowedImageTypes[http.DetectContentType(head[:n])]
	if !ok {
		writeError(w, http.StatusUnsupportedMediaType, "only JPG, PNG, WEBP or GIF images are allowed")
		return
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		writeStoreError(w, err)
		return
	}

	sub := time.Now().Format("2006/01")
	dir := filepath.Join(h.uploadDir, sub)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		writeStoreError(w, err)
		return
	}
	name := randomName() + ext
	dst, err := os.Create(filepath.Join(dir, name))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	defer dst.Close()
	if _, err := io.Copy(dst, file); err != nil {
		writeStoreError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"url": "/uploads/" + sub + "/" + name})
}

func randomName() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// noListingFS serves files but never directory listings.
type noListingFS struct{ fs http.FileSystem }

func (n noListingFS) Open(name string) (http.File, error) {
	f, err := n.fs.Open(name)
	if err != nil {
		return nil, err
	}
	if st, err := f.Stat(); err == nil && st.IsDir() {
		f.Close()
		return nil, os.ErrNotExist
	}
	return f, nil
}

// spaHandler serves the built frontend and falls back to index.html for
// client-side routes.
func spaHandler(dir string) http.Handler {
	files := http.FileServer(noListingFS{http.Dir(dir)})
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := filepath.Join(dir, filepath.Clean("/"+r.URL.Path))
		if st, err := os.Stat(path); err == nil && !st.IsDir() {
			if strings.HasPrefix(r.URL.Path, "/assets/") {
				w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			}
			files.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Cache-Control", "no-cache")
		http.ServeFile(w, r, filepath.Join(dir, "index.html"))
	})
}
