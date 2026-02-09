package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jay6909/dino-internal-wallet-service/internal/data/repository"
	data_requests "github.com/jay6909/dino-internal-wallet-service/internal/data/requests"
	"github.com/jay6909/dino-internal-wallet-service/internal/enums"
	"github.com/jay6909/dino-internal-wallet-service/internal/utils"
	"gorm.io/gorm"
)

type WalletHandler struct {
	walletRepository repository.WalletRepository
	userRepository   repository.UserRepository
}

func NewWalletHandler(walletRepository repository.WalletRepository, userRepository repository.UserRepository) *WalletHandler {
	return &WalletHandler{walletRepository: walletRepository, userRepository: userRepository}
}

func (h *WalletHandler) RegisterRoutes(r *gin.RouterGroup) {
	route := r.Group("/wallets")
	route.POST("/top-up", h.TopUp)
	route.POST("/spend", h.Spend)
	route.GET("/balance", h.GetWalletByOwner)
	route.POST("/bonus/:id", h.Bonus)

}

func (h *WalletHandler) Bonus(c *gin.Context) {
	req := &data_requests.BonusRequest{}
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	transaction, err := h.walletRepository.GetTransactionByIdempotencyKey(req.IdempotencyKey)

	if err == nil && transaction != nil {
		// already processed
		c.JSON(http.StatusOK, gin.H{
			"message":     "Bonus added (idempotent)",
			"transaction": transaction,
		})
		return
	}

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		// real DB error
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	systemWallet, err := h.walletRepository.GetSystemWalletByCurrencyType(req.CurrencyTypeID.String())
	if status, err := utils.CheckGormError(c, err); err != nil {
		c.AbortWithStatusJSON(status, gin.H{"error": err})
		return
	}
	wallet, err := h.CheckUserWalletIfNotCreate(req.OwnerID, req.CurrencyTypeID)
	if utils.ReturnIfGormError(c, err) {
		return
	}

	//from system wallet to user wallet
	if err := h.walletRepository.Transfer(systemWallet.ID.String(), wallet.ID.String(),
		req.CurrencyTypeID.String(), req.IdempotencyKey, req.Amount, enums.TransactionTypeBonus); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Bonus added",
	})
}

func (h *WalletHandler) TopUp(c *gin.Context) {
	req := &data_requests.TopUpRequest{}
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	transaction, err := h.walletRepository.GetTransactionByIdempotencyKey(req.IdempotencyKey)

	if err == nil && transaction != nil {
		// already processed
		c.JSON(http.StatusOK, gin.H{
			"message":     "Top-up successful (idempotent)",
			"transaction": transaction,
		})
		return
	}

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		// real DB error
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	systemWallet, err := h.walletRepository.GetSystemWalletByCurrencyType(req.CurrencyTypeID.String())
	if utils.ReturnIfGormError(c, err) {
		return
	}
	wallet, err := h.CheckUserWalletIfNotCreate(req.OwnerID, req.CurrencyTypeID)
	if utils.ReturnIfGormError(c, err) {
		return
	}

	if wallet == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get or create user wallet",
		})
		return
	}

	//from system wallet to user wallet
	if err := h.walletRepository.Transfer(systemWallet.ID.String(), wallet.ID.String(),
		req.CurrencyTypeID.String(), req.IdempotencyKey, req.Amount, enums.TransactionTypeTopUp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Top-up successful",
	})
}

func (h *WalletHandler) Spend(c *gin.Context) {
	req := &data_requests.SpendRequest{}
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	transaction, err := h.walletRepository.GetTransactionByIdempotencyKey(req.IdempotencyKey)

	if err == nil && transaction != nil {
		// already processed
		c.JSON(http.StatusOK, gin.H{
			"message": "Spend successful (idempotent)",
			// "transaction": transaction,
		})
		return
	}

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		// real DB error
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	systemWallet, err := h.walletRepository.GetSystemWalletByCurrencyType(req.CurrencyTypeID.String())
	if utils.ReturnIfGormError(c, err) {
		return
	}
	wallet, err := h.CheckUserWalletIfNotCreate(req.OwnerID, req.CurrencyTypeID)
	if utils.ReturnIfGormError(c, err) {
		return
	}

	//from user wallet to system wallet
	if err := h.walletRepository.Transfer(wallet.ID.String(), systemWallet.ID.String(),
		req.CurrencyTypeID.String(), req.IdempotencyKey, req.Amount, enums.TransactionTypeSpend); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Spend successful",
		"wallet":  wallet,
	})

}

func (h *WalletHandler) CheckUserWalletIfNotCreate(ownerID, currencyTypeID uuid.UUID) (*repository.Wallet, error) {
	owner, err := h.userRepository.GetUserByID(ownerID.String())
	if err != nil {
		return nil, err
	}

	wallet, err := h.walletRepository.GetWalletByOwner(owner.Role, ownerID.String(), currencyTypeID.String())
	if status, err := utils.CheckGormError(nil, err); err != nil {
		if status == http.StatusNotFound {
			var newWallet = &repository.Wallet{
				ID:             uuid.New(),
				OwnerType:      owner.Role,
				OwnerID:        ownerID,
				CurrencyTypeID: currencyTypeID,
				Balance:        0,
			}
			if err := h.walletRepository.CreateWallet(newWallet); err != nil {
				return nil, err
			}
			return newWallet, nil
		} else {

			return nil, err
		}
	}
	return wallet, nil
}

func (h *WalletHandler) GetWalletByOwner(c *gin.Context) {
	ownerType := c.Param("owner_type")
	ownerID := c.Query("owner_id")
	currencyTypeID := c.Query("currency_type_id")

	if ownerType == "" || ownerID == "" || currencyTypeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "owner_type, owner_id, and currency_type_id are required in query params",
		})
		return
	}
	wallet, err := h.walletRepository.GetWalletByOwner(ownerType, ownerID, currencyTypeID)
	if utils.ReturnIfGormError(c, err) {
		return
	}
	c.JSON(200, wallet)
}
