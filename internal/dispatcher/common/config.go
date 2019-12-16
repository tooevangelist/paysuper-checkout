package common

import "time"

type Config struct {
	CookieDomain string `envconfig:"COOKIE_DOMAIN" required:"true"`
	AllowOrigin  string `envconfig:"ALLOW_ORIGIN" default:"*"`

	PaymentFormJsLibraryUrl string `envconfig:"PAYMENT_FORM_JS_LIBRARY_URL" required:"true"`
	OrderInlineFormUrlMask  string `envconfig:"ORDER_INLINE_FORM_URL_MASK" required:"true"`
	WebsocketUrl            string `envconfig:"WEBSOCKET_URL" default:"wss://cf.tst.protocol.one/connection/websocket"`

	CustomerTokenCookiesLifetime time.Duration `envconfig:"CUSTOMER_TOKEN_COOKIES_LIFETIME" default:"2592000"`
}
