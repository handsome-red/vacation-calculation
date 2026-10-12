package ports

import (
	"context"

	"github.com/handsome-red/vacation-calculation/internal/domain/vacation"
)

type ReportExporter interface {
	Export(ctx context.Context, report vacation.Report) ([]byte, error)
}
