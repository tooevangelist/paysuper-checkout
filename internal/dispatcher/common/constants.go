package common

import (
	"github.com/ProtocolONE/go-core/v2/pkg/logger"
	"github.com/paysuper/paysuper-billing-server/pkg"
)

type Dictionary struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

const (
	RequestParameterId        = "id"
	RequestParameterOrderId   = "order_id"
	RequestParameterZipUsa    = "zip_usa"
	RequestParameterReceiptId = "receipt_id"

	OrderFieldProjectId     = "PP_PROJECT_ID"
	OrderFieldSignature     = "PP_SIGNATURE"
	OrderFieldAmount        = "PP_AMOUNT"
	OrderFieldCurrency      = "PP_CURRENCY"
	OrderFieldAccount       = "PP_ACCOUNT"
	OrderFieldOrderId       = "PP_ORDER_ID"
	OrderFieldPaymentMethod = "PP_PAYMENT_METHOD"
	OrderFieldUrlVerify     = "PP_URL_VERIFY"
	OrderFieldUrlNotify     = "PP_URL_NOTIFY"
	OrderFieldUrlSuccess    = "PP_URL_SUCCESS"
	OrderFieldUrlFail       = "PP_URL_FAIL"
	OrderFieldPayerEmail    = "PP_PAYER_EMAIL"
	OrderFieldPayerPhone    = "PP_PAYER_PHONE"
	OrderFieldDescription   = "PP_DESCRIPTION"
	OrderFieldRegion        = "PP_REGION"

	QueryParameterNameUtmMedium   = "utm_medium"
	QueryParameterNameUtmCampaign = "utm_campaign"
	QueryParameterNameUtmSource   = "utm_source"

	ErrorMessageMask = "field validation for '%s' failed on the '%s' tag"

	HeaderAcceptLanguage      = "Accept-Language"
	HeaderUserAgent           = "User-Agent"
	HeaderXApiSignatureHeader = "X-API-SIGNATURE"
	HeaderReferer             = "referer"

	CustomerTokenCookiesName = "_ps_ctkn"

	ErrorFieldService = "service"
	ErrorFieldMethod  = "method"
	ErrorFieldRequest = "request"

	InternalErrorTemplate = "internal error"

	ValidationParameterOrderId   = "OrderId"
	ValidationParameterOrderUuid = "OrderUuid"
)

var (
	OrderReservedWords = map[string]bool{
		OrderFieldProjectId:     true,
		OrderFieldSignature:     true,
		OrderFieldAmount:        true,
		OrderFieldCurrency:      true,
		OrderFieldAccount:       true,
		OrderFieldOrderId:       true,
		OrderFieldDescription:   true,
		OrderFieldPaymentMethod: true,
		OrderFieldUrlVerify:     true,
		OrderFieldUrlNotify:     true,
		OrderFieldUrlSuccess:    true,
		OrderFieldUrlFail:       true,
		OrderFieldPayerEmail:    true,
		OrderFieldPayerPhone:    true,
		OrderFieldRegion:        true,
	}
)

func LogSrvCallFailedGRPC(log logger.Logger, err error, name, method string, req interface{}) {
	log.Error(pkg.ErrorGrpcServiceCallFailed,
		logger.PairArgs(
			ErrorFieldService, name,
			ErrorFieldMethod, method,
		),
		logger.WithPrettyFields(logger.Fields{"err": err, ErrorFieldRequest: req}),
	)
}
