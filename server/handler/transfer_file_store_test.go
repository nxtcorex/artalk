package handler

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTransferFileStoreExpiresUnconsumedFile(t *testing.T) {
	store := newTransferFileStore(10 * time.Millisecond)
	defer store.Close()

	filename, err := store.Save(strings.NewReader("pending import"), 1024)
	require.NoError(t, err)
	require.FileExists(t, filename)

	require.Eventually(t, func() bool {
		_, err := os.Stat(filename)
		return os.IsNotExist(err)
	}, time.Second, 10*time.Millisecond)
}

func TestTransferFileStoreRejectsOversizedFile(t *testing.T) {
	store := newTransferFileStore(time.Hour)
	defer store.Close()

	filename, err := store.Save(strings.NewReader("too large"), 3)
	require.ErrorIs(t, err, errTransferFileTooLarge)
	require.Empty(t, filename)
}
