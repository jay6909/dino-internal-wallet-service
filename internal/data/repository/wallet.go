package repository

import (
	"errors"
	"sort"

	"github.com/google/uuid"
	"github.com/jay6909/dino-internal-wallet-service/internal/enums"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Wallet struct {
	ID             uuid.UUID `gorm:"type:char(36);primaryKey"`
	OwnerType      string    `gorm:"type:varchar(32);not null;uniqueIndex:uniq_owner_currency,priority:1"`
	OwnerID        uuid.UUID `gorm:"type:char(36);not null;uniqueIndex:uniq_owner_currency,priority:2"`
	CurrencyTypeID uuid.UUID `gorm:"type:char(36);not null;uniqueIndex:uniq_owner_currency,priority:3"`
	Balance        int64     `gorm:"not null;default:0;check:balance >= 0"`
	Version        int       `gorm:"not null;default:0"`
	BaseTimeStamps
}

type WalletTransaction struct {
	ID              uuid.UUID `gorm:"type:char(36);primaryKey"`
	WalletID        uuid.UUID `gorm:"type:char(36);not null;uniqueIndex:uniq_wallet_idempotency,priority:1"`
	TransactionType string    `gorm:"type:varchar(32);not null"`
	Amount          int64     `gorm:"not null"`
	BalanceAfter    int64     `gorm:"not null"`
	ReferenceID     string    `gorm:"type:varchar(64);not null;index"`
	IdempotencyKey  string    `gorm:"type:varchar(64);not null;uniqueIndex:uniq_wallet_idempotency,priority:2"`
	BaseTimeStamps
}

type WalletRepository interface {
	GetWalletByOwner(ownerType string, ownerID string, currencyTypeID string) (*Wallet, error)
	GetSystemWalletByCurrencyType(currencyTypeID string) (*Wallet, error)
	CreateWallet(wallet *Wallet) error
	Transfer(fromWalletID, toWalletID, currencyTypeID, idempotencyKey string, amount int64, transactionType enums.TransactionType) error
	GetTransactionByIdempotencyKey(idempotencyKey string) (*WalletTransaction, error)
}

type walletRepositoryImpl struct {
	db *gorm.DB
}

func NewWalletRepository(db *gorm.DB) WalletRepository {
	return &walletRepositoryImpl{db: db}
}

func (w *walletRepositoryImpl) CreateWallet(wallet *Wallet) error {
	return w.db.Create(wallet).Error
}
func (w *walletRepositoryImpl) GetSystemWalletByCurrencyType(currencyTypeID string) (*Wallet, error) {
	var wallet Wallet
	if err := w.db.Where("owner_type = ? AND currency_type_id = ?", "system", currencyTypeID).First(&wallet).Error; err != nil {
		return nil, err
	}
	return &wallet, nil

}

// GetWalletByOwner implements WalletRepository.
func (w *walletRepositoryImpl) GetWalletByOwner(ownerType string, ownerID string, currencyTypeID string) (*Wallet, error) {
	var wallet Wallet
	if err := w.db.Where("owner_type = ? AND owner_id = ? AND currency_type_id = ?", ownerType, ownerID, currencyTypeID).First(&wallet).Error; err != nil {
		return nil, err
	}
	return &wallet, nil
}

// GetTransactionByIdempotencyKey implements WalletRepository.
func (w *walletRepositoryImpl) GetTransactionByIdempotencyKey(idempotencyKey string) (*WalletTransaction, error) {
	var transaction WalletTransaction
	if err := w.db.Where("idempotency_key = ?", idempotencyKey).First(&transaction).Error; err != nil {
		return nil, err
	}
	return &transaction, nil
}

// UpdateWalletBalance implements WalletRepository.
func (w *walletRepositoryImpl) Transfer(fromWalletID, toWalletID, currencyTypeID, idempotencyKey string, amount int64, transactionType enums.TransactionType) error {
	return w.db.Transaction(func(tx *gorm.DB) error {
		// Lock the wallet record for update
		var fromWallet Wallet
		var toWallet Wallet
		if transactionType != "" {
			transactionType = enums.TransactionTypeTopUp
		}

		walletIDs := []string{fromWalletID, toWalletID}
		sort.Strings(walletIDs)
		wallets := make(map[string]*Wallet)
		for _, id := range walletIDs {
			var wlt Wallet
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("id = ? AND currency_type_id = ?",
					id, currencyTypeID).First(&wlt).Error; err != nil {
				return err
			}
			wallets[id] = &wlt
		}
		toWallet = *wallets[toWalletID]
		fromWallet = *wallets[fromWalletID]

		if fromWallet.Balance < amount {
			return errors.New("insufficient balance")
		}

		if fromWallet.Balance-amount < 0 {
			return errors.New("treasury has insufficient balance")
		}

		if fromWallet.ID == toWallet.ID {
			return errors.New("cannot transfer to the same wallet")
		}
		key, err := w.GetTransactionByIdempotencyKey(idempotencyKey)
		if err == nil && key != nil {
			// If a transaction with the same idempotency key exists, return it without creating a new one
			return nil
		}
		referenceId := uuid.New().String()
		debit := WalletTransaction{
			ID:              uuid.New(),
			WalletID:        fromWallet.ID,
			TransactionType: enums.TransactionTypeDebit,
			Amount:          -amount,
			BalanceAfter:    fromWallet.Balance - amount,
			ReferenceID:     referenceId,
			IdempotencyKey:  idempotencyKey,
		}
		credit := WalletTransaction{
			ID:              uuid.New(),
			WalletID:        toWallet.ID,
			TransactionType: string(transactionType),
			Amount:          amount,
			BalanceAfter:    toWallet.Balance + amount,
			ReferenceID:     referenceId,
			IdempotencyKey:  idempotencyKey,
		}
		if err := tx.Create(&debit).Error; err != nil {
			return err
		}

		if err := tx.Create(&credit).Error; err != nil {
			return err
		}
		if err := tx.Model(&fromWallet).Update("balance", fromWallet.Balance-amount).Error; err != nil {
			return err
		}

		if err := tx.Model(&toWallet).Update("balance", toWallet.Balance+amount).Error; err != nil {
			return err
		}

		return nil // commit
	})
}
