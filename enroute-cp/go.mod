module github.com/saarasio/enroute/enroute-cp

go 1.17

require (
	github.com/labstack/echo/v4 v4.6.1
	github.com/pkg/errors v0.9.1
	github.com/saarasio/enroute/enroute-cp/docs v0.7.0
	github.com/saarasio/enroute/enroute-dp v0.7.0
	github.com/sirupsen/logrus v1.6.0
	github.com/stretchr/testify v1.6.1
	github.com/swaggo/echo-swagger v1.1.0
)

require (
	github.com/KyleBanks/depth v1.2.1 // indirect
	github.com/PuerkitoBio/purell v1.1.1 // indirect
	github.com/PuerkitoBio/urlesc v0.0.0-20170810143723-de5bf2ad4578 // indirect
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/census-instrumentation/opencensus-proto v0.2.1 // indirect
	github.com/cncf/xds/go v0.0.0-20210312221358-fbca930ec8ed // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/envoyproxy/go-control-plane v0.9.9 // indirect
	github.com/envoyproxy/protoc-gen-validate v0.1.0 // indirect
	github.com/evanphx/json-patch v4.9.0+incompatible // indirect
	github.com/go-logr/logr v0.4.0 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/jsonreference v0.19.5 // indirect
	github.com/go-openapi/spec v0.20.0 // indirect
	github.com/go-openapi/swag v0.19.12 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang-jwt/jwt v3.2.2+incompatible // indirect
	github.com/golang/protobuf v1.4.3 // indirect
	github.com/google/go-cmp v0.5.2 // indirect
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/googleapis/gnostic v0.4.1 // indirect
	github.com/hashicorp/golang-lru v0.5.1 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.10 // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.3 // indirect
	github.com/labstack/gommon v0.3.0 // indirect
	github.com/mailru/easyjson v0.7.6 // indirect
	github.com/mattn/go-colorable v0.1.8 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_golang v1.1.0 // indirect
	github.com/prometheus/client_model v0.0.0-20190812154241-14fe0d1b01d4 // indirect
	github.com/prometheus/common v0.6.0 // indirect
	github.com/prometheus/procfs v0.0.3 // indirect
	github.com/swaggo/files v0.0.0-20190704085106-630677cd5c14 // indirect
	github.com/swaggo/swag v1.7.0 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.1 // indirect
	golang.org/x/crypto v0.0.0-20210817164053-32db794688a5 // indirect
	golang.org/x/net v0.0.0-20210913180222-943fd674d43e // indirect
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d // indirect
	golang.org/x/sys v0.0.0-20210910150752-751e447fb3d0 // indirect
	golang.org/x/term v0.0.0-20210220032956-6a3ed077a48d // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/time v0.0.0-20210220033141-f8bda1e9f3ba // indirect
	golang.org/x/tools v0.0.0-20210106214847-113979e3529a // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	google.golang.org/appengine v1.6.5 // indirect
	google.golang.org/genproto v0.0.0-20200526211855-cb27e3aa2013 // indirect
	google.golang.org/grpc v1.36.0 // indirect
	google.golang.org/protobuf v1.25.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776 // indirect
	k8s.io/api v0.21.0 // indirect
	k8s.io/apimachinery v0.21.0 // indirect
	k8s.io/client-go v0.21.0 // indirect
	k8s.io/klog/v2 v2.8.0 // indirect
	k8s.io/utils v0.0.0-20201110183641-67b214c5f920 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.1.0 // indirect
	sigs.k8s.io/yaml v1.2.0 // indirect
)

replace github.com/saarasio/enroute/enroute-cp/docs => ./docs

replace github.com/saarasio/enroute/enroute-dp => ../enroute-dp
