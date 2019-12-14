package handlers

import (
	"context"
	"github.com/ProtocolONE/go-core/v2/pkg/config"
	"github.com/ProtocolONE/go-core/v2/pkg/logger"
	"github.com/ProtocolONE/go-core/v2/pkg/provider"
	"github.com/paysuper/paysuper-checkout/internal/dispatcher/common"
	"github.com/stretchr/testify/assert"
	"gopkg.in/go-playground/validator.v9"
	"testing"
)

func Test_Provider_Ok(t *testing.T) {
	handlers, fn, err := ProviderHandlers(
		config.Initial{},
		common.Services{},
		&validator.Validate{},
		provider.AwareSet{Logger: logger.NewMock(context.Background(), &logger.Config{}, true)},
		&common.Config{},
	)

	asserts := assert.New(t)
	asserts.True(len(handlers) > 0)
	asserts.IsType(func() {}, fn)
	asserts.NoError(err)
}
