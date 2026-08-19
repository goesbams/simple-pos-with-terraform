package handler

import (
	"net/http"

	"github.com/goesbams/simple-pos-with-terraform/internal/model"
	"github.com/goesbams/simple-pos-with-terraform/internal/service"
	"github.com/labstack/echo/v4"
)

type ProductHandler struct {
	productService service.ProductService
}

func NewProductHandler(productService service.ProductService) *ProductHandler {
	return &ProductHandler{productService: productService}
}

func (h *ProductHandler) GetAllProducts(c echo.Context) error {
	products, err := h.productService.GetAllProducts(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": products,
	})
}

func (h *ProductHandler) GetProductByID(c echo.Context) error {
	id := c.Param("id")
	p, err := h.productService.GetProductByID(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "product not found"})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": p,
	})
}

func (h *ProductHandler) CreateProduct(c echo.Context) error {
	var req model.CreateProductRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	if req.Name == "" || req.Price <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "name and positive price required"})
	}

	p, err := h.productService.CreateProduct(c.Request().Context(), req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message": "product created successfully",
		"data":    p,
	})
}

func (h *ProductHandler) UpdateProduct(c echo.Context) error {
	id := c.Param("id")
	var req model.UpdateProductRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	p, err := h.productService.UpdateProduct(c.Request().Context(), id, req)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "product updated successfully",
		"data":    p,
	})
}

func (h *ProductHandler) SoftDeleteProduct(c echo.Context) error {
	id := c.Param("id")
	if err := h.productService.DeleteProduct(c.Request().Context(), id); err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "product deleted successfully (soft delete)",
	})
}
