package api

import (
	"context"
	"time"

	"github.com/dagu-org/dagu/api/v2"
	"github.com/dagu-org/dagu/internal/build"
	"github.com/dagu-org/dagu/internal/frontend/metrics"
	"github.com/dagu-org/dagu/internal/stringutil"
)

func (a *API) GetHealthStatus(_ context.Context, _ api.GetHealthStatusRequestObject) (api.GetHealthStatusResponseObject, error) {
	return &api.GetHealthStatus200JSONResponse{
		Status:    api.HealthResponseStatusHealthy,
		Version:   build.Version,
		Uptime:    int(metrics.GetUptime()),
		Timestamp: stringutil.FormatTime(time.Now()),
	}, nil
}
