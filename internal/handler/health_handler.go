package handler

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

type HealthHandler struct {
	db *sql.DB
}

func NewHealthHandler(db *sql.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

func (h *HealthHandler) HealthCheck(c echo.Context) error {
	dbStatus := "up"
	if h.db != nil {
		if err := h.db.PingContext(c.Request().Context()); err != nil {
			dbStatus = "down"
		}
	} else {
		dbStatus = "uninitialized"
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":    "healthy",
		"database":  dbStatus,
		"service":   "Simple POS API",
		"timestamp": time.Now(),
	})
}
