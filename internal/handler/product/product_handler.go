package product

import (
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/idalgo-2021/product-catalog-backend/internal/domain"
	handlers "github.com/idalgo-2021/product-catalog-backend/internal/handler"
)

type ProductHandler struct {
	service      ProductService
	defaultLimit int
	maxLimit     int
	logger       *zap.Logger
}

func NewProductHandler(service ProductService, defaultLimit, maxLimit int, logger *zap.Logger) *ProductHandler {
	return &ProductHandler{
		service:      service,
		defaultLimit: defaultLimit,
		maxLimit:     maxLimit,
		logger:       logger,
	}
}

// GetProduct godoc
// @Summary      Get a product by ID
// @Description  Get detailed information about a product
// @Tags         product
// @Accept       json
// @Produce      json
// @Param        product_id path string true "Product ID"
// @Success      200  {object}  handler.Response{data=ProductResponse}
// @Failure      400  {object}  handler.Response{error=handler.ErrorInfo}
// @Failure      404  {object}  handler.Response{error=handler.ErrorInfo}
// @Failure      500  {object}  handler.Response{error=handler.ErrorInfo}
// @Router       /products/{product_id} [get]
func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["product_id"]

	product, err := h.service.GetProduct(r.Context(), id)
	if err != nil {
		handlers.HandleError(w, h.logger, "failed to get product", err)
		return
	}

	productDTO := toProductResponse(product)

	handlers.Success(w, http.StatusOK, productDTO)
}

func toProductResponse(p *domain.Product) ProductResponse {
	imgs := make([]ImageResponse, 0, len(p.Images))

	for _, item := range p.Images {
		imgs = append(imgs, ImageResponse{
			URL:    item.URL,
			Width:  item.Width,
			Height: item.Height,
		})
	}

	attrs := make(map[string]string, len(p.Attributes))
	for k, v := range p.Attributes {
		attrs[k] = v
	}

	return ProductResponse{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
		Currency:    p.Currency,
		Available:   p.Available,
		Thumbnail: ImageResponse{
			URL:    p.Thumbnail.URL,
			Width:  p.Thumbnail.Width,
			Height: p.Thumbnail.Height,
		},
		Images:     imgs,
		Attributes: attrs,
		CreatedAt:  p.CreatedAt,
		UpdatedAt:  p.UpdatedAt,
	}
}

// ListProducts godoc
// @Summary      List products
// @Description  Get a paginated list of products
// @Tags         product
// @Accept       json
// @Produce      json
// @Param        limit query int false "Limit"
// @Param        offset query int false "Offset"
// @Success      200  {object}  handler.Response{data=[]ProductResponse}
// @Failure      500  {object}  handler.Response{error=handler.ErrorInfo}
// @Router       /products [get]
func (h *ProductHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	limit, offset := handlers.ParsePagination(r, h.defaultLimit, h.maxLimit)

	products, total, err := h.service.ListProducts(r.Context(), limit, offset)
	if err != nil {
		handlers.HandleError(w, h.logger, "failed to list products", err)
		return
	}

	productsDTO := toProductsResponse(products)

	handlers.SuccessWithMeta(w, http.StatusOK, productsDTO, &handlers.Meta{
		Limit:  limit,
		Offset: offset,
		Total:  total,
	})
}

func toProductsResponse(products []domain.Product) []ProductResponse {
	result := make([]ProductResponse, 0, len(products))

	for i := range products {
		result = append(result, toProductResponse(&products[i]))
	}

	return result
}
