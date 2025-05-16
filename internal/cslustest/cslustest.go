package cslustest

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
	cslusDownloader      download.Downloader
	cslusDownloaderSetup sync.Once
)

func FindEntity(tb testing.TB, entityID string) search.Entity[search.Value] {
	tb.Helper()

	logger := log.NewTestLogger()

	cslusDownloaderSetup.Do(func() {
		pkg, err := fshelp.FindPkgDir()
		require.NoError(tb, err)

		conf := download.Config{
			InitialDataDirectory: filepath.Join(pkg, "sources", "csl_us", "testdata"),
		}
		conf.IncludedLists = append(conf.IncludedLists, search.SourceUSCSL)

		dl, err := download.NewDownloader(logger, conf)
		require.NoError(tb, err)

		cslusDownloader = dl
	})

	stats, err := cslusDownloader.RefreshAll(context.Background())
	require.NoError(tb, err)

	for _, entity := range stats.Entities {
		if entity.SourceID == entityID && entity.Source == search.SourceUSCSL {
			return entity
		}
	}

	tb.Fatalf("US CSL entity %s not found", entityID)

	return search.Entity[search.Value]{}
}
