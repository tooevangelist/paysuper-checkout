package handlers

import (
	"github.com/ProtocolONE/go-core/v2/pkg/logger"
	"github.com/ProtocolONE/go-core/v2/pkg/provider"
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"github.com/paysuper/paysuper-checkout/internal/dispatcher/common"
	"github.com/paysuper/paysuper-checkout/internal/helpers"
	"net/http"
)

const (
	removeSavedCardPath = "/saved_card"
)

type RecurringRoute struct {
	dispatch common.HandlerSet
	cfg      common.Config
	provider.LMT
}

func NewRecurringRoute(set common.HandlerSet, cfg *common.Config) *RecurringRoute {
	set.AwareSet.Logger = set.AwareSet.Logger.WithFields(logger.Fields{"router": "RecurringRoute"})
	return &RecurringRoute{
		dispatch: set,
		LMT:      &set.AwareSet,
		cfg:      *cfg,
	}
}

func (h *RecurringRoute) Route(groups *common.Groups) {
	groups.Common.DELETE(removeSavedCardPath, h.removeSavedCard)
}

func (h *RecurringRoute) removeSavedCard(ctx echo.Context) error {
	req := &grpc.DeleteSavedCardRequest{
		Cookie: helpers.GetRequestCookie(ctx, common.CustomerTokenCookiesName),
	}

	if err := h.dispatch.BindAndValidate(req, ctx); err != nil {
		return err
	}

	res, err := h.dispatch.Services.Billing.DeleteSavedCard(ctx.Request().Context(), req)

	if err != nil {
		return h.dispatch.SrvCallHandler(req, err, pkg.ServiceName, "DeleteSavedCard")
	}

	if res.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.NoContent(http.StatusOK)
}
