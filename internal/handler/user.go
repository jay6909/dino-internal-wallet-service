package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/jay6909/dino-internal-wallet-service/internal/data/repository"
	"github.com/jay6909/dino-internal-wallet-service/internal/utils"
)

type UserHandler struct {
	userRepository repository.UserRepository
}

func NewUserHandler(userRepository repository.UserRepository) *UserHandler {

	return &UserHandler{userRepository: userRepository}
}

func (h *UserHandler) RegisterRoutes(r *gin.Engine) {
	route := r.Group("/users")
	route.GET("/:id", h.GetUserByID)

}
func (h *UserHandler) GetUserByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(400, gin.H{"error": "id is required"})
		return
	}
	user, err := h.userRepository.GetUserByID(id)
	if utils.ReturnIfGormError(c, err) {
		return
	}
	c.JSON(200, user)
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var user repository.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if err := h.userRepository.CreateUser(&user); err != nil {
		if utils.ReturnIfGormError(c, err) {
			return
		}
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
}
