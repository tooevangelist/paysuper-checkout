package handlers

import (
	"github.com/ProtocolONE/go-core/v2/pkg/config"
	"github.com/ProtocolONE/go-core/v2/pkg/provider"
	"github.com/paysuper/paysuper-checkout/internal/dispatcher/common"
	"gopkg.in/go-playground/validator.v9"
)

func ProviderHandlers(
	initial config.Initial,
	srv common.Services,
	validator *validator.Validate,
	set provider.AwareSet,
	cfg *common.Config,
) (common.Handlers, func(), error) {
	hSet := common.HandlerSet{
		Services: srv,
		Validate: validator,
		AwareSet: set,
	}
	copyCfg := *cfg

	return []common.Handler{
		NewCountryRoute(hSet, &copyCfg),
		NewOrderRoute(hSet, &copyCfg),
		NewPaymentRoute(hSet, &copyCfg),
		NewRecurringRoute(hSet, &copyCfg),
	}, func() {}, nil
}
