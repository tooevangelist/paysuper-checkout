package handlers

import (
	"context"
	"github.com/ProtocolONE/go-core/v2/pkg/logger"
	"github.com/ProtocolONE/go-core/v2/pkg/provider"
	u "github.com/PuerkitoBio/purell"
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/billing"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"github.com/paysuper/paysuper-checkout/internal/dispatcher/common"
	"github.com/paysuper/paysuper-checkout/internal/helpers"
	"net/http"
	"time"
)

const (
	orderPath                = "/order"
	orderIdPath              = "/order/:order_id"
	orderReCreatePath        = "/order/recreate"
	orderLanguagePath        = "/orders/:order_id/language"
	orderCustomerPath        = "/orders/:order_id/customer"
	orderBillingAddressPath  = "/orders/:order_id/billing_address"
	orderNotifySalesPath     = "/orders/:order_id/notify_sale"
	orderNotifyNewRegionPath = "/orders/:order_id/notify_new_region"
	orderPlatformPath        = "/orders/:order_id/platform"
	orderReceiptPath         = "/orders/receipt/:receipt_id/:order_id"
	paylinkIdPath            = "/paylink/:id"
)

const (
	errorTemplateName = "error.html"
)

type CreateOrderJsonProjectResponse struct {
	Id              string                    `json:"id"`
	PaymentFormUrl  string                    `json:"payment_form_url"`
	PaymentFormData *grpc.PaymentFormJsonData `json:"payment_form_data,omitempty"`
}

type ListOrdersRequest struct {
	MerchantId    string   `json:"merchant_id" validate:"required,hexadecimal,len=24"`
	FileType      string   `json:"file_type" validate:"required"`
	Template      string   `json:"template" validate:"omitempty,hexadecimal"`
	Id            string   `json:"id" validate:"omitempty,uuid"`
	Project       []string `json:"project" validate:"omitempty,dive,hexadecimal,len=24"`
	PaymentMethod []string `json:"payment_method" validate:"omitempty,dive,hexadecimal,len=24"`
	Country       []string `json:"country" validate:"omitempty,dive,alpha,len=2"`
	Status        []string `json:"status," validate:"omitempty,dive,alpha,oneof=created processed canceled rejected refunded chargeback pending"`
	PmDateFrom    int64    `json:"pm_date_from" validate:"omitempty,numeric,gt=0"`
	PmDateTo      int64    `json:"pm_date_to" validate:"omitempty,numeric,gt=0"`
}

type OrderRoute struct {
	dispatch common.HandlerSet
	cfg      *common.Config
	provider.LMT
}

func NewOrderRoute(set common.HandlerSet, cfg *common.Config) *OrderRoute {
	set.AwareSet.Logger = set.AwareSet.Logger.WithFields(logger.Fields{"router": "OrderRoute"})
	return &OrderRoute{
		dispatch: set,
		LMT:      &set.AwareSet,
		cfg:      cfg,
	}
}

func (h *OrderRoute) Route(groups *common.Groups) {
	groups.Common.POST(orderPath, h.createJson)
	groups.Common.GET(orderIdPath, h.getPaymentFormData)
	groups.Common.POST(orderReCreatePath, h.recreateOrder)
	groups.Common.PATCH(orderLanguagePath, h.changeLanguage)
	groups.Common.PATCH(orderCustomerPath, h.changeCustomer)
	groups.Common.POST(orderBillingAddressPath, h.processBillingAddress)
	groups.Common.POST(orderNotifySalesPath, h.notifySale)
	groups.Common.POST(orderNotifyNewRegionPath, h.notifyNewRegion)
	groups.Common.POST(orderPlatformPath, h.changePlatform)
	groups.Common.GET(orderReceiptPath, h.getReceipt)
	groups.Common.GET(paylinkIdPath, h.getOrderForPaylink)
}

func (h *OrderRoute) createJson(ctx echo.Context) error {
	req := &billing.OrderCreateRequest{}

	if err := (&common.OrderJsonBinder{}).Bind(req, ctx); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	req.Cookie = helpers.GetRequestCookie(ctx, common.CustomerTokenCookiesName)

	if err := h.dispatch.Validate.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	// If request contain user object then paysuper must check request signature
	if req.User != nil {
		httpErr := common.CheckProjectAuthRequestSignature(h.dispatch, ctx, req.ProjectId)

		if httpErr != nil {
			return httpErr
		}
	}

	ctxReq := ctx.Request().Context()
	req.IssuerUrl = ctx.Request().Header.Get(common.HeaderReferer)

	var (
		order *billing.Order
	)

	// If request contain prepared order identifier than try to get order by this identifier
	if req.PspOrderUuid != "" {
		req := &grpc.IsOrderCanBePayingRequest{
			OrderId:   req.PspOrderUuid,
			ProjectId: req.ProjectId,
		}
		rsp, err := h.dispatch.Services.Billing.IsOrderCanBePaying(ctxReq, req)

		if err != nil {
			return h.dispatch.SrvCallHandler(req, err, pkg.ServiceName, "IsOrderCanBePaying")
		}

		if rsp.Status != pkg.ResponseStatusOk {
			return echo.NewHTTPError(int(rsp.Status), rsp.Message)
		}

		order = rsp.Item
	} else {
		rsp, err := h.dispatch.Services.Billing.OrderCreateProcess(ctxReq, req)

		if err != nil {
			return h.dispatch.SrvCallHandler(req, err, pkg.ServiceName, "OrderCreateProcess")
		}

		if rsp.Status != http.StatusOK {
			return echo.NewHTTPError(int(rsp.Status), rsp.Message)
		}

		order = rsp.Item
	}

	response := &CreateOrderJsonProjectResponse{
		Id:             order.Uuid,
		PaymentFormUrl: h.cfg.OrderInlineFormUrlMask + "?order_id=" + order.Uuid,
	}

	return ctx.JSON(http.StatusOK, response)
}

func (h *OrderRoute) getPaymentFormData(ctx echo.Context) error {
	req := &grpc.PaymentFormJsonDataRequest{
		Locale:  ctx.Request().Header.Get(common.HeaderAcceptLanguage),
		Ip:      ctx.RealIP(),
		Referer: ctx.Request().Header.Get(common.HeaderReferer),
		Cookie:  helpers.GetRequestCookie(ctx, common.CustomerTokenCookiesName),
	}

	if err := h.dispatch.BindAndValidate(req, ctx); err != nil {
		return err
	}

	res, err := h.dispatch.Services.Billing.PaymentFormJsonDataProcess(ctx.Request().Context(), req)

	if err != nil {
		return h.dispatch.SrvCallHandler(req, err, pkg.ServiceName, "PaymentFormJsonDataProcess")
	}

	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	expire := time.Now().Add(time.Duration(h.cfg.CustomerTokenCookiesLifetimeHours) * time.Hour)
	helpers.SetResponseCookie(ctx, common.CustomerTokenCookiesName, res.Cookie, h.cfg.CookieDomain, expire)

	return ctx.JSON(http.StatusOK, res.Item)
}

func (h *OrderRoute) recreateOrder(ctx echo.Context) error {
	req := &grpc.OrderReCreateProcessRequest{}

	if err := h.dispatch.BindAndValidate(req, ctx); err != nil {
		return err
	}

	res, err := h.dispatch.Services.Billing.OrderReCreateProcess(ctx.Request().Context(), req)

	if err != nil {
		return h.dispatch.SrvCallHandler(req, err, pkg.ServiceName, "OrderReCreateProcess")
	}

	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	order := res.Item
	response := &CreateOrderJsonProjectResponse{
		Id:             order.Uuid,
		PaymentFormUrl: h.cfg.OrderInlineFormUrlMask + "?order_id=" + order.Uuid,
	}

	return ctx.JSON(http.StatusOK, response)
}

func (h *OrderRoute) changeLanguage(ctx echo.Context) error {
	req := &grpc.PaymentFormUserChangeLangRequest{
		AcceptLanguage: ctx.Request().Header.Get(common.HeaderAcceptLanguage),
		UserAgent:      ctx.Request().Header.Get(common.HeaderUserAgent),
		Ip:             ctx.RealIP(),
	}

	if err := h.dispatch.BindAndValidate(req, ctx); err != nil {
		return err
	}

	res, err := h.dispatch.Services.Billing.PaymentFormLanguageChanged(ctx.Request().Context(), req)

	if err != nil {
		return h.dispatch.SrvCallHandler(req, err, pkg.ServiceName, "PaymentFormLanguageChanged")
	}

	if res.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Item)
}

func (h *OrderRoute) changeCustomer(ctx echo.Context) error {
	req := &grpc.PaymentFormUserChangePaymentAccountRequest{
		AcceptLanguage: ctx.Request().Header.Get(common.HeaderAcceptLanguage),
		UserAgent:      ctx.Request().Header.Get(common.HeaderUserAgent),
		Ip:             ctx.RealIP(),
	}

	if err := h.dispatch.BindAndValidate(req, ctx); err != nil {
		return err
	}

	res, err := h.dispatch.Services.Billing.PaymentFormPaymentAccountChanged(ctx.Request().Context(), req)

	if err != nil {
		return h.dispatch.SrvCallHandler(req, err, pkg.ServiceName, "PaymentFormPaymentAccountChanged")
	}

	if res.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Item)
}

func (h *OrderRoute) processBillingAddress(ctx echo.Context) error {
	req := &grpc.ProcessBillingAddressRequest{
		Cookie: helpers.GetRequestCookie(ctx, common.CustomerTokenCookiesName),
		Ip:     ctx.RealIP(),
	}

	if err := h.dispatch.BindAndValidate(req, ctx); err != nil {
		return err
	}

	res, err := h.dispatch.Services.Billing.ProcessBillingAddress(ctx.Request().Context(), req)

	if err != nil {
		return h.dispatch.SrvCallHandler(req, err, pkg.ServiceName, "ProcessBillingAddress")
	}

	if res.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	expire := time.Now().Add(time.Duration(h.cfg.CustomerTokenCookiesLifetimeHours) * time.Hour)
	helpers.SetResponseCookie(ctx, common.CustomerTokenCookiesName, res.Cookie, h.cfg.CookieDomain, expire)

	return ctx.JSON(http.StatusOK, res.Item)
}

func (h *OrderRoute) notifySale(ctx echo.Context) error {
	req := &grpc.SetUserNotifyRequest{}

	if err := h.dispatch.BindAndValidate(req, ctx); err != nil {
		return err
	}

	_, err := h.dispatch.Services.Billing.SetUserNotifySales(ctx.Request().Context(), req)

	if err != nil {
		return h.dispatch.SrvCallHandler(req, err, pkg.ServiceName, "SetUserNotifySales")
	}

	return ctx.NoContent(http.StatusNoContent)
}

func (h *OrderRoute) notifyNewRegion(ctx echo.Context) error {
	req := &grpc.SetUserNotifyRequest{}

	if err := h.dispatch.BindAndValidate(req, ctx); err != nil {
		return err
	}

	_, err := h.dispatch.Services.Billing.SetUserNotifyNewRegion(ctx.Request().Context(), req)

	if err != nil {
		return h.dispatch.SrvCallHandler(req, err, pkg.ServiceName, "SetUserNotifyNewRegion")
	}

	return ctx.NoContent(http.StatusNoContent)
}

func (h *OrderRoute) changePlatform(ctx echo.Context) error {
	req := &grpc.PaymentFormUserChangePlatformRequest{}

	if err := h.dispatch.BindAndValidate(req, ctx); err != nil {
		return err
	}

	res, err := h.dispatch.Services.Billing.PaymentFormPlatformChanged(ctx.Request().Context(), req)

	if err != nil {
		return h.dispatch.SrvCallHandler(req, err, pkg.ServiceName, "PaymentFormPlatformChanged")
	}

	if res.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Item)
}

func (h *OrderRoute) getReceipt(ctx echo.Context) error {
	req := &grpc.OrderReceiptRequest{}

	if err := h.dispatch.BindAndValidate(req, ctx); err != nil {
		return err
	}

	res, err := h.dispatch.Services.Billing.OrderReceipt(ctx.Request().Context(), req)

	if err != nil {
		return h.dispatch.SrvCallHandler(req, err, pkg.ServiceName, "OrderReceipt")
	}

	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Receipt)
}

func (h *OrderRoute) getOrderForPaylink(ctx echo.Context) error {
	paylinkId := ctx.Param(common.RequestParameterId)

	go func() {
		req := &grpc.PaylinkRequestById{Id: paylinkId}
		// call with background context to prevent request abandoning when redirect will bw returned in response below
		_, err := h.dispatch.Services.Billing.IncrPaylinkVisits(context.Background(), req)

		if err != nil {
			common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "IncrPaylinkVisits", req)
		}
	}()

	qParams := ctx.QueryParams()

	req := &billing.OrderCreateByPaylink{
		PaylinkId:   paylinkId,
		PayerIp:     ctx.RealIP(),
		IssuerUrl:   ctx.Request().Header.Get(common.HeaderReferer),
		UtmSource:   qParams.Get(common.QueryParameterNameUtmSource),
		UtmMedium:   qParams.Get(common.QueryParameterNameUtmMedium),
		UtmCampaign: qParams.Get(common.QueryParameterNameUtmCampaign),
		IsEmbedded:  false,
		Cookie:      helpers.GetRequestCookie(ctx, common.CustomerTokenCookiesName),
	}

	res, err := h.dispatch.Services.Billing.OrderCreateByPaylink(ctx.Request().Context(), req)

	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "OrderCreateByPaylink", req)
		return ctx.Render(http.StatusBadRequest, errorTemplateName, map[string]interface{}{})
	}

	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	qParams.Set("order_id", res.Item.Uuid)
	inlineFormRedirectUrl, err := u.NormalizeURLString(
		h.cfg.OrderInlineFormUrlMask+"?"+qParams.Encode(),
		u.FlagsUsuallySafeGreedy|u.FlagRemoveDuplicateSlashes,
	)

	if err != nil {
		h.L().Error("NormalizeURLString failed", logger.PairArgs("err", err.Error()))
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorUnknown)
	}

	return ctx.Redirect(http.StatusFound, inlineFormRedirectUrl)
}
