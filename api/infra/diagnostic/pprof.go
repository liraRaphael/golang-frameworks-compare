package diagnostic

import (
	"net/http"
	_ "net/http/pprof"
	"github.com/liraraphael/go-framework-bench/api/infra/observability/logger"
)

func StartPprof(addr string) {
	l := logger.GetLogger()
	l.Info("Starting pprof server", logger.LoggerFieldType{"address": addr})
	go func() {
		if err := http.ListenAndServe(addr, nil); err != nil {
			l.Error("pprof server failed", err, nil)
		}
	}()
}
