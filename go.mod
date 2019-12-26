module github.com/paysuper/paysuper-checkout

require (
	github.com/Jeffail/gabs v1.1.1 // indirect
	github.com/ProtocolONE/go-core/v2 v2.1.0
	github.com/PuerkitoBio/purell v1.1.1
	github.com/SAP/go-hdb v0.13.2 // indirect
	github.com/SermoDigital/jose v0.9.2-0.20161205224733-f6df55f235c2 // indirect
	github.com/alexeyco/simpletable v0.0.0-20190222165044-2eb48bcee7cf
	github.com/asaskevich/govalidator v0.0.0-20180720115003-f9ffefc3facf // indirect
	github.com/elazarl/go-bindata-assetfs v1.0.0 // indirect
	github.com/fatih/color v1.7.0
	github.com/fatih/structs v1.1.0 // indirect
	github.com/globalsign/mgo v0.0.0-20181015135952-eeefdecb41b8
	github.com/go-log/log v0.1.0
	github.com/google/uuid v1.1.1
	github.com/google/wire v0.3.0
	github.com/gurukami/typ/v2 v2.0.1
	github.com/hashicorp/go-memdb v0.0.0-20181108192425-032f93b25bec // indirect
	github.com/hashicorp/go-plugin v0.0.0-20181212150838-f444068e8f5a // indirect
	github.com/keybase/go-crypto v0.0.0-20181127160227-255a5089e85a // indirect
	github.com/labstack/echo/v4 v4.1.11
	github.com/micro/go-micro v1.8.0
	github.com/micro/go-plugins v1.2.0
	github.com/micro/go-rcache v0.2.1 // indirect
	github.com/micro/util v0.2.0 // indirect
	github.com/mitchellh/copystructure v1.0.0 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
	github.com/paysuper/paysuper-billing-server v1.1.1-0.20191225165128-842136ef7d03
	github.com/pkg/errors v0.8.1
	github.com/spf13/cobra v0.0.5
	github.com/stretchr/testify v1.4.0
	github.com/ttacon/builder v0.0.0-20170518171403-c099f663e1c2 // indirect
	github.com/ttacon/libphonenumber v1.0.1
	go.uber.org/automaxprocs v1.2.0
	gopkg.in/go-playground/validator.v9 v9.29.1
)

replace (
	github.com/gogo/protobuf v0.0.0-20190410021324-65acae22fc9 => github.com/gogo/protobuf v1.2.2-0.20190723190241-65acae22fc9d
	github.com/hashicorp/consul => github.com/hashicorp/consul v1.5.1
	github.com/hashicorp/consul/api => github.com/hashicorp/consul/api v1.1.0
)

go 1.12
