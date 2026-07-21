package handler

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

var errTransferFileTooLarge = errors.New("transfer import file is too large")

const (
	transferFilePrefix = "artalk-import-file-"
	transferFileTTL    = time.Hour
)

type transferFileStore struct {
	mu    sync.Mutex
	ttl   time.Duration
	files map[string]*time.Timer
}

func newTransferFileStore(ttl time.Duration) *transferFileStore {
	return &transferFileStore{
		ttl:   ttl,
		files: map[string]*time.Timer{},
	}
}

func (s *transferFileStore) Save(src io.Reader, maxBytes int64) (string, error) {
	tmpFile, err := os.CreateTemp("", transferFilePrefix)
	if err != nil {
		return "", fmt.Errorf("create temporary import file: %w", err)
	}

	filename := tmpFile.Name()
	keepFile := false
	defer func() {
		_ = tmpFile.Close()
		if !keepFile {
			_ = os.Remove(filename)
		}
	}()

	written, err := io.Copy(tmpFile, io.LimitReader(src, maxBytes+1))
	if err != nil {
		return "", fmt.Errorf("write temporary import file: %w", err)
	}
	if written > maxBytes {
		return "", fmt.Errorf("%w: limit is %d bytes", errTransferFileTooLarge, maxBytes)
	}
	if err := tmpFile.Close(); err != nil {
		return "", fmt.Errorf("close temporary import file: %w", err)
	}

	keepFile = true
	s.track(filename)
	return filename, nil
}

func (s *transferFileStore) Take(filename string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	timer, ok := s.files[filename]
	if !ok {
		return "", false
	}

	delete(s.files, filename)
	timer.Stop()
	return filename, true
}

func (s *transferFileStore) Discard(filename string) {
	path, ok := s.Take(filename)
	if ok {
		_ = os.Remove(path)
	}
}

func (s *transferFileStore) Close() {
	s.mu.Lock()
	files := s.files
	s.files = map[string]*time.Timer{}
	s.mu.Unlock()

	for filename, timer := range files {
		timer.Stop()
		_ = os.Remove(filename)
	}
}

func (s *transferFileStore) track(filename string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var timer *time.Timer
	timer = time.AfterFunc(s.ttl, func() {
		s.mu.Lock()
		defer s.mu.Unlock()

		if current, ok := s.files[filename]; ok && current == timer {
			delete(s.files, filename)
			_ = os.Remove(filename)
		}
	})
	s.files[filename] = timer
}
