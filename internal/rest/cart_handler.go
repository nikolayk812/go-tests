package rest

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/nikolayk812/go-tests/internal/rest/mapper"
	"github.com/nikolayk812/go-tests/internal/service"
	"github.com/nikolayk812/go-tests/pkg/dto"
	"net/http"
)

type CartHandler struct {
	service service.CartService
}

func NewCartHandler(service service.CartService) (*CartHandler, error) {
	if service == nil {
		return nil, errors.New("service is nil")
	}

	return &CartHandler{service: service}, nil
}

func (h *CartHandler) GetCart(c *gin.Context) {
	ownerID := c.Param("owner_id")
	if ownerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "owner_id is empty"})
		return
	}

	cart, err := h.service.GetCart(c, ownerID)
	if err != nil {
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "An unexpected error occurred"})
		return
	}

	cartDTO := mapper.CartToDTO(cart)

	c.JSON(http.StatusOK, cartDTO)
}

func (h *CartHandler) AddItem(c *gin.Context) {
	ownerID := c.Param("owner_id")
	if ownerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "owner_id is empty"})
		return
	}

	var itemDTO dto.CartItem
	if err := c.BindJSON(&itemDTO); err != nil {
		_ = c.Error(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot parse request body"})
		return
	}

	item, err := mapper.CartItemFromDTO(itemDTO)
	if err != nil {
		_ = c.Error(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err := h.service.AddItem(c, ownerID, item); err != nil {
		_ = c.Error(err)

		if errors.Is(err, service.ErrCartDuplicateItem) {
			c.JSON(http.StatusConflict, gin.H{"error": "item already exists in the cart"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "An unexpected error occurred"})
		return
	}

	c.Status(http.StatusCreated)
}
