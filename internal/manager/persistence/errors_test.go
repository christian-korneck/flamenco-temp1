// SPDX-License-Identifier: GPL-3.0-or-later
package persistence

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestNotFoundErrors(t *testing.T) {
	assert.ErrorIs(t, ErrJobNotFound, gorm.ErrRecordNotFound)
	assert.ErrorIs(t, ErrTaskNotFound, gorm.ErrRecordNotFound)

	assert.Contains(t, ErrJobNotFound.Error(), "job")
	assert.Contains(t, ErrTaskNotFound.Error(), "task")
}

func TestTranslateGormJobError(t *testing.T) {
	assert.Nil(t, translateGormJobError(nil))
	assert.Equal(t, ErrJobNotFound, translateGormJobError(gorm.ErrRecordNotFound))

	otherError := errors.New("this error is not special for this function")
	assert.Equal(t, otherError, translateGormJobError(otherError))
}

func TestTranslateGormTaskError(t *testing.T) {
	assert.Nil(t, translateGormTaskError(nil))
	assert.Equal(t, ErrTaskNotFound, translateGormTaskError(gorm.ErrRecordNotFound))

	otherError := errors.New("this error is not special for this function")
	assert.Equal(t, otherError, translateGormTaskError(otherError))
}
