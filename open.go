package levelqueue

import (
	lediscfg "github.com/ledisdb/ledisdb/config"
	"github.com/ledisdb/ledisdb/ledis"
)

func Open(dir string) (*ledis.Ledis, error) {
	cfg := lediscfg.NewConfigDefault()
	cfg.DataDir = dir

	return ledis.Open(cfg)
}
