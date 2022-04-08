package alerts

import (
	"context"
	"stocks/models"
)

type Alert = string

type AlertParser interface {
	GetAlerts(ctx context.Context, holdingsMap map[string]models.Holding) ([]Alert, error)
}
