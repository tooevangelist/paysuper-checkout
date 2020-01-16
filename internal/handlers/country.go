package handlers

import (
	"github.com/ProtocolONE/go-core/v2/pkg/logger"
	"github.com/ProtocolONE/go-core/v2/pkg/provider"
	"github.com/labstack/echo/v4"
	billing "github.com/paysuper/paysuper-proto/go/billingpb"
	"github.com/paysuper/paysuper-checkout/internal/dispatcher/common"
	"net/http"
)

const (
	paymentCountriesOrderIdPath = "/payment_countries/:order_id"
)

type CountryRoute struct {
	dispatch common.HandlerSet
	cfg      *common.Config
	provider.LMT
}

func NewCountryRoute(set common.HandlerSet, cfg *common.Config) *CountryRoute {
	set.AwareSet.Logger = set.AwareSet.Logger.WithFields(logger.Fields{"router": "CountryRoute"})
	return &CountryRoute{
		dispatch: set,
		LMT:      &set.AwareSet,
		cfg:      cfg,
	}
}

func (h *CountryRoute) Route(groups *common.Groups) {
	groups.Common.GET(paymentCountriesOrderIdPath, h.getPaymentCountries)
}

func (h *CountryRoute) getPaymentCountries(ctx echo.Context) error {
	req := &billing.GetCountriesListForOrderRequest{}

	if err := h.dispatch.BindAndValidate(req, ctx); err != nil {
		return err
	}

	res, err := h.dispatch.Services.Billing.GetCountriesListForOrder(ctx.Request().Context(), req)

	if err != nil {
		return h.dispatch.SrvCallHandler(req, err, billing.ServiceName, "GetCountriesListForOrder")
	}

	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Item)
}
