package data_requests

import "github.com/google/uuid"

type BonusRequest struct {
	IdempotencyKey string    `json:"idempotency_key" binding:"required"`
	OwnerID        uuid.UUID `json:"owner_id" binding:"required"`
	CurrencyTypeID uuid.UUID `json:"currency_type_id" binding:"required"`
	Amount         int64     `json:"amount" binding:"required,gt=0"`
}
type TopUpRequest struct {
	IdempotencyKey string    `json:"idempotency_key" binding:"required"`
	OwnerID        uuid.UUID `json:"owner_id" binding:"required"`
	CurrencyTypeID uuid.UUID `json:"currency_type_id" binding:"required"`
	Amount         int64     `json:"amount" binding:"required,gt=0"`
}

type SpendRequest struct {
	IdempotencyKey string    `json:"idempotency_key" binding:"required"`
	OwnerID        uuid.UUID `json:"owner_id" binding:"required"`
	CurrencyTypeID uuid.UUID `json:"currency_type_id" binding:"required"`
	Amount         int64     `json:"amount" binding:"required,gt=0"`
}
