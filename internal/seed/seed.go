package seed

import (
	"errors"

	"github.com/google/uuid"
	config_db "github.com/jay6909/dino-internal-wallet-service/internal/config/db"
	config_env "github.com/jay6909/dino-internal-wallet-service/internal/config/env"
	"github.com/jay6909/dino-internal-wallet-service/internal/data/repository"
	"gorm.io/gorm"
)

func SeedDb() {
	appEnv, err := config_env.LoadAppEnv()
	if err != nil {
		panic(err)
	}
	database, err := config_db.NewGormDB(appEnv.DatabaseConfig.DSN)
	if err != nil {
		panic(err)
	}
	systemUser := SeedSystemUser(database)
	currencyTypes := SeedCurrencyTypes(database)
	SeedUsers(database, currencyTypes)

	SeedTreasury(database, *systemUser, currencyTypes)
}

func SeedUsers(db config_db.DB, currencyTypes []repository.CurrencyType) {
	uuid1, err := uuid.Parse("d3f57c3b-3a35-4b6a-9c22-3f8f9e3c1111")
	if err != nil {
		panic("failed to parse uuid for user 1")
	}
	uuid2, err := uuid.Parse("d3f57c3b-3a35-4b6a-9c22-3f8f9e3c2222")
	if err != nil {
		panic("failed to parse uuid for user 2")
	}

	users := []repository.User{
		{ID: uuid1, Name: "Alice", Role: "user"},
		{ID: uuid2, Name: "Bob", Role: "user"},
	}

	// 1️⃣ Seed users (idempotent)
	for _, user := range users {
		var existing repository.User
		err := db.GetDB().
			Where("id = ?", user.ID).
			First(&existing).Error

		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			panic(err)
		}

		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := db.GetDB().Create(&user).Error; err != nil {
				panic(err)
			}
		}
	}

	// 2️⃣ Seed wallets (3 per user)
	for _, user := range users {
		for _, ct := range currencyTypes {

			var wallet repository.Wallet
			err := db.GetDB().
				Where(
					"owner_type = ? AND owner_id = ? AND currency_type_id = ?",
					user.Role,
					user.ID,
					ct.ID,
				).
				First(&wallet).Error

			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				panic(err)
			}

			if errors.Is(err, gorm.ErrRecordNotFound) {
				newWallet := repository.Wallet{
					ID:             uuid.New(),
					OwnerType:      user.Role,
					OwnerID:        user.ID,
					CurrencyTypeID: ct.ID,
					Balance:        1000,
				}

				if err := db.GetDB().Create(&newWallet).Error; err != nil {
					panic(err)
				}
			}
		}
	}
}

func SeedSystemUser(db config_db.DB) *repository.User {
	systemUser := &repository.User{}
	err := db.GetDB().Where("name = ? AND role = ?",
		"system",
		"system").First(&systemUser).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		panic(err)
	}
	if err == nil {
		return systemUser
	}
	systemUser = &repository.User{
		ID:   uuid.New(),
		Name: "system",
		Role: "system",
	}
	err = db.GetDB().Create(systemUser).Error
	if err != nil {
		panic(err)
	}

	return systemUser

}
func SeedCurrencyTypes(db config_db.DB) []repository.CurrencyType {
	names := []string{"gold", "diamond", "loyalty_points"}
	var result []repository.CurrencyType

	for _, name := range names {
		ct := &repository.CurrencyType{}
		err := db.GetDB().
			Where("name = ?", name).
			First(&ct).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			panic(err)
		}
		if err == nil {
			result = append(result, *ct)
			continue
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ct = &repository.CurrencyType{
				ID:   uuid.New(),
				Name: name,
			}
			err = db.GetDB().Create(ct).Error
			if err != nil {
				panic(err)
			}
			result = append(result, *ct)

		}

	}
	return result

}
func SeedTreasury(
	db config_db.DB,
	systemUser repository.User,
	currencyTypes []repository.CurrencyType,
) {
	for _, ct := range currencyTypes {
		wallet := repository.Wallet{
			ID:             uuid.New(),
			OwnerType:      "system",
			OwnerID:        systemUser.ID,
			CurrencyTypeID: ct.ID,
			Balance:        1_000_000,
		}

		err := db.GetDB().
			Where(
				"owner_type = ? AND owner_id = ? AND currency_type_id = ?",
				"system", systemUser.ID, ct.ID,
			).
			First(&wallet).Error
		if err == nil && wallet.ID != uuid.Nil {
			continue
		}
		err = db.GetDB().Create(&wallet).Error

		if err != nil {
			panic("failed to seed treasury wallet")
		}
	}
}
