package common

import "time"

type Config struct {
	CookieDomain string `envconfig:"COOKIE_DOMAIN" required:"true"`
	AllowOrigin  string `envconfig:"ALLOW_ORIGIN" default:"*"`

	PaymentFormJsLibraryUrl string `envconfig:"PAYMENT_FORM_JS_LIBRARY_URL" required:"true"`
	// OrderInlineFormUrlMask url like a https://checkout.tst.pay.super.com/pay/order/
	OrderInlineFormUrlMask string `envconfig:"ORDER_INLINE_FORM_URL_MASK" required:"true"`

	CustomerTokenCookiesLifetime time.Duration `envconfig:"CUSTOMER_TOKEN_COOKIES_LIFETIME" default:"2592000"`
}
