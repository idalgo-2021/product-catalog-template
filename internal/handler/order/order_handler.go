package order

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/idalgo-2021/product-catalog-backend/internal/domain"
	handlers "github.com/idalgo-2021/product-catalog-backend/internal/handler"
	"github.com/idalgo-2021/product-catalog-backend/internal/middleware"
)

type OrderHandler struct {
	service      OrderService
	defaultLimit int
	maxLimit     int
	logger       *zap.Logger
}

func NewOrderHandler(service OrderService, defaultLimit, maxLimit int, logger *zap.Logger) *OrderHandler {
	return &OrderHandler{
		service:      service,
		defaultLimit: defaultLimit,
		maxLimit:     maxLimit,
		logger:       logger,
	}
}

type OrderItemRequest struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type CreateOrderRequest struct {
	Items []OrderItemRequest `json:"items"`
}

type UpdateOrderRequest struct {
	Status *string             `json:"status,omitempty"`
	Items  *[]OrderItemRequest `json:"items,omitempty"`
}

// ListOrders godoc
// @Summary      List orders
// @Description  Get a paginated list of user orders
// @Tags         order
// @Accept       json
// @Produce      json
// @Param        limit query int false "Limit"
// @Param        offset query int false "Offset"
// @Success      200  {object}  handler.Response{data=[]OrderResponse}
// @Failure      401  {object}  handler.Response{error=handler.ErrorInfo}
// @Failure      500  {object}  handler.Response{error=handler.ErrorInfo}
// @Security     BearerAuth
// @Router       /orders [get]
func (h *OrderHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		handlers.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	limit, offset := handlers.ParsePagination(r, h.defaultLimit, h.maxLimit)

	orders, total, err := h.service.ListOrders(r.Context(), userID, limit, offset)
	if err != nil {
		handlers.HandleError(w, h.logger, "failed to list orders", err)
		return
	}

	ordersDTO := toOrdersResponse(orders)

	handlers.SuccessWithMeta(w, http.StatusOK, ordersDTO, &handlers.Meta{
		Limit:  limit,
		Offset: offset,
		Total:  total,
	})
}

func toOrdersResponse(orders []domain.Order) []OrderResponse {
	if len(orders) == 0 {
		return []OrderResponse{}
	}

	result := make([]OrderResponse, 0, len(orders))

	for _, o := range orders {
		result = append(result, toOrderResponse(&o))
	}

	return result
}

// GetOrder godoc
// @Summary      Get an order by ID
// @Description  Get detailed information about a specific order
// @Tags         order
// @Accept       json
// @Produce      json
// @Param        order_id path string true "Order ID"
// @Success      200  {object}  handler.Response{data=OrderResponse}
// @Failure      401  {object}  handler.Response{error=handler.ErrorInfo}
// @Failure      403  {object}  handler.Response{error=handler.ErrorInfo}
// @Failure      404  {object}  handler.Response{error=handler.ErrorInfo}
// @Failure      500  {object}  handler.Response{error=handler.ErrorInfo}
// @Security     BearerAuth
// @Router       /orders/{order_id} [get]
func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["order_id"]

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		handlers.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	order, err := h.service.GetOrder(r.Context(), id, userID)
	if err != nil {
		handlers.HandleError(w, h.logger, "failed to get order", err)
		return
	}

	orderDTO := toOrderResponse(order)

	handlers.Success(w, http.StatusOK, orderDTO)
}

func toOrderResponse(o *domain.Order) OrderResponse {
	items := make([]OrderItemResponse, 0, len(o.Items))

	for _, item := range o.Items {
		items = append(items, OrderItemResponse{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
		})
	}

	return OrderResponse{
		ID:        o.ID,
		Status:    string(o.Status),
		Total:     o.Total,
		Currency:  o.Currency,
		CreatedAt: o.CreatedAt,
		UpdatedAt: o.UpdatedAt,
		Items:     items,
	}
}

// CreateOrder godoc
// @Summary      Create a new order
// @Description  Create a new order with items
// @Tags         order
// @Accept       json
// @Produce      json
// @Param        input body CreateOrderRequest true "Order details"
// @Success      201  {object}  handler.Response{data=CreateOrderResponse}
// @Failure      400  {object}  handler.Response{error=handler.ErrorInfo}
// @Failure      401  {object}  handler.Response{error=handler.ErrorInfo}
// @Failure      500  {object}  handler.Response{error=handler.ErrorInfo}
// @Security     BearerAuth
// @Router       /orders [post]
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handlers.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		handlers.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	domainItems := make([]domain.OrderItem, len(req.Items))
	for i, item := range req.Items {
		domainItems[i] = domain.OrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}
	}

	order, err := h.service.CreateOrder(r.Context(), userID, domainItems)
	if err != nil {
		handlers.HandleError(w, h.logger, "failed to create order", err)
		return
	}

	// Отдельный маппер и отдельная DTO
	orderDTO := toCreateOrderResponse(order)

	handlers.Success(w, http.StatusCreated, orderDTO)
}

func toCreateOrderResponse(o *domain.Order) CreateOrderResponse {
	items := make([]OrderItemResponse, 0, len(o.Items))

	for _, item := range o.Items {
		items = append(items, OrderItemResponse{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
		})
	}

	return CreateOrderResponse{
		ID:        o.ID,
		Status:    string(o.Status),
		Total:     o.Total,
		Currency:  o.Currency,
		Items:     items,
		CreatedAt: o.CreatedAt,
	}
}

// UpdateOrder godoc
// @Summary      Update an order
// @Description  Update the status or items of an existing order
// @Tags         order
// @Accept       json
// @Produce      json
// @Param        order_id path string true "Order ID"
// @Param        input body UpdateOrderRequest true "Order update details"
// @Success      200  {object}  handler.Response{data=OrderResponse}
// @Failure      400  {object}  handler.Response{error=handler.ErrorInfo}
// @Failure      401  {object}  handler.Response{error=handler.ErrorInfo}
// @Failure      403  {object}  handler.Response{error=handler.ErrorInfo}
// @Failure      404  {object}  handler.Response{error=handler.ErrorInfo}
// @Failure      500  {object}  handler.Response{error=handler.ErrorInfo}
// @Security     BearerAuth
// @Router       /orders/{order_id} [patch]
func (h *OrderHandler) UpdateOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["order_id"]

	var req UpdateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handlers.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		handlers.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	input := domain.UpdateOrderInput{}
	if req.Status != nil {
		status := domain.OrderStatus(*req.Status)
		input.Status = &status
	}
	if req.Items != nil {
		domainItems := make([]domain.OrderItem, len(*req.Items))
		for i, item := range *req.Items {
			domainItems[i] = domain.OrderItem{
				ProductID: item.ProductID,
				Quantity:  item.Quantity,
			}
		}
		input.Items = &domainItems
	}

	order, err := h.service.UpdateOrder(r.Context(), id, userID, input)
	if err != nil {
		handlers.HandleError(w, h.logger, "failed to update order", err)
		return
	}

	orderDTO := toOrderResponse(order)

	handlers.Success(w, http.StatusOK, orderDTO)
}
