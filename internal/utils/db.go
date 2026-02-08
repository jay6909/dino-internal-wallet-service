package utils

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CheckGormError checks if the error is a GORM err and returns appropreate status code and error message if it is
func CheckGormError(c *gin.Context, err error) (int, error) {
	if err == nil {
		return 0, nil
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return http.StatusNotFound, fmt.Errorf("not found")
	}
	if errors.Is(err, gorm.ErrInvalidData) {
		return http.StatusBadRequest, fmt.Errorf("invalid data")
	}
	if errors.Is(err, gorm.ErrInvalidField) {
		return http.StatusBadRequest, fmt.Errorf("invalid field")
	}
	if errors.Is(err, gorm.ErrUnsupportedRelation) {
		return http.StatusBadRequest, fmt.Errorf("unsupported relation")
	}
	if errors.Is(err, gorm.ErrPrimaryKeyRequired) {
		return http.StatusBadRequest, fmt.Errorf("primary key required")
	}
	if errors.Is(err, gorm.ErrModelValueRequired) {
		return http.StatusBadRequest, fmt.Errorf("model value required")
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return http.StatusConflict, fmt.Errorf("duplicated key")
	}
	if errors.Is(err, gorm.ErrEmptySlice) {
		return http.StatusBadRequest, fmt.Errorf("empty slice")
	}

	return http.StatusInternalServerError, err
}

// ReturnIfGormError checks if the error is a GORM err found and aborts with json response and  appropreate status code if it is
func ReturnIfGormError(c *gin.Context, err error) {
	status, errNew := CheckGormError(c, err)
	if status != 0 && errNew != nil {
		c.AbortWithStatusJSON(status, gin.H{"error": errNew.Error()})
		return
	}
	if errNew == nil && status == 0 {

		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
