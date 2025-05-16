package ofactest

import (
	"context"
	"path/filepath"
	"sync"
	"testing"

	"github.com/moov-io/base/log"
	"github.com/moov-io/watchman/internal/download"
	"github.com/moov-io/watchman/internal/fshelp"
	"github.com/moov-io/watchman/pkg/search"

	"github.com/stretchr/testify/require"
)

var (
	ofacDownloader      download.Downloader
	ofacDownloaderSetup sync.Once
)

func GetDownloader(tb testing.TB) download.Downloader {
	tb.Helper()

	logger := log.NewTestLogger()

	ofacDownloaderSetup.Do(func() {
		pkg, err := fshelp.FindPkgDir()
		require.NoError(tb, err)

		initialDir := filepath.Join(pkg, "..", "internal", "ofactest", "testdata")

		// By default use a small set of OFAC records, but use the larger record set for benchmarks
		if _, ok := tb.(*testing.B); ok {
			initialDir = filepath.Join(pkg, "sources", "ofac", "testdata")
		}

		conf := download.Config{
			InitialDataDirectory: initialDir,
		}
		conf.IncludedLists = append(conf.IncludedLists, search.SourceUSOFAC)

		dl, err := download.NewDownloader(logger, conf)
		require.NoError(tb, err)

		ofacDownloader = dl
	})

	return ofacDownloader
}

func FindEntity(tb testing.TB, entityID string) search.Entity[search.Value] {
	tb.Helper()

	stats, err := GetDownloader(tb).RefreshAll(context.Background())
	require.NoError(tb, err)

	for _, entity := range stats.Entities {
		if entity.SourceID == entityID && entity.Source == search.SourceUSOFAC {
			return entity
		}
	}

	tb.Fatalf("OFAC entity %s not found", entityID)

	return search.Entity[search.Value]{}
}
