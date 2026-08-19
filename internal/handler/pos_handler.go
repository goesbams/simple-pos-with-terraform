package handler

import (
	"net/http"

	"github.com/goesbams/simple-pos-with-terraform/internal/model"
	"github.com/goesbams/simple-pos-with-terraform/internal/service"
	"github.com/labstack/echo/v4"
)

type POSHandler struct {
	posService service.POSService
}

func NewPOSHandler(posService service.POSService) *POSHandler {
	return &POSHandler{posService: posService}
}

func (h *POSHandler) Checkout(c echo.Context) error {
	var req model.CheckoutRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	if req.CartID == "" {
		req.CartID = "default_cart"
	}
	if req.PaymentType == "" {
		req.PaymentType = "cash"
	}

	txn, err := h.posService.Checkout(c.Request().Context(), req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "checkout transaction initiated",
		"data":    txn,
	})
}

func (h *POSHandler) GetTransactionByID(c echo.Context) error {
	id := c.Param("id")
	txn, err := h.posService.GetTransactionByID(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "transaction not found"})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": txn,
	})
}

func (h *POSHandler) GetAllTransactions(c echo.Context) error {
	transactions, err := h.posService.GetAllTransactions(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": transactions,
	})
}

func (h *POSHandler) MidtransNotificationWebhook(c echo.Context) error {
	var payload model.MidtransNotificationPayload
	if err := c.Bind(&payload); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid notification payload"})
	}

	if err := h.posService.ProcessMidtransWebhook(c.Request().Context(), payload); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "notification processed successfully",
	})
}
