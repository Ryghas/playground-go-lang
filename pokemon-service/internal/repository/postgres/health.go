package postgres

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// DBHealthChecker mengecek koneksi database untuk readiness probe.
type DBHealthChecker struct {
	db *gorm.DB
}

func NewHealthChecker(db *gorm.DB) *DBHealthChecker {
	return &DBHealthChecker{db: db}
}

func (h *DBHealthChecker) Ping(ctx context.Context) error {
	sqlDB, err := h.db.DB()
	if err != nil {
		return fmt.Errorf("DBHealthChecker.Ping get sql.DB: %w", err)
	}
	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("DBHealthChecker.Ping: %w", err)
	}
	return nil
}
