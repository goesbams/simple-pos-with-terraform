package handler

import (
	"net/http"

	"github.com/goesbams/simple-pos-with-terraform/internal/model"
	"github.com/goesbams/simple-pos-with-terraform/internal/service"
	"github.com/labstack/echo/v4"
)

type CartHandler struct {
	cartService service.CartService
}

func NewCartHandler(cartService service.CartService) *CartHandler {
	return &CartHandler{cartService: cartService}
}

func (h *CartHandler) GetCart(c echo.Context) error {
	cartID := c.QueryParam("cart_id")
	if cartID == "" {
		cartID = "default_cart"
	}

	cart, err := h.cartService.GetCart(c.Request().Context(), cartID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": cart,
	})
}

func (h *CartHandler) AddToCart(c echo.Context) error {
	cartID := c.QueryParam("cart_id")
	if cartID == "" {
		cartID = "default_cart"
	}

	var req model.AddToCartRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	cart, err := h.cartService.AddToCart(c.Request().Context(), cartID, req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "item added to cart",
		"data":    cart,
	})
}

func (h *CartHandler) RemoveFromCart(c echo.Context) error {
	cartID := c.QueryParam("cart_id")
	if cartID == "" {
		cartID = "default_cart"
	}
	productID := c.Param("product_id")

	cart, err := h.cartService.RemoveFromCart(c.Request().Context(), cartID, productID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "item removed from cart",
		"data":    cart,
	})
}

func (h *CartHandler) ClearCart(c echo.Context) error {
	cartID := c.QueryParam("cart_id")
	if cartID == "" {
		cartID = "default_cart"
	}

	if err := h.cartService.ClearCart(c.Request().Context(), cartID); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "cart cleared successfully",
	})
}
