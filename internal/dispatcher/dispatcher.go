package dispatcher

import (
	"context"
	"github.com/ProtocolONE/go-core/v2/pkg/invoker"
	"github.com/ProtocolONE/go-core/v2/pkg/logger"
	"github.com/ProtocolONE/go-core/v2/pkg/provider"
	"github.com/alexeyco/simpletable"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/paysuper/paysuper-checkout/internal/dispatcher/common"
	"github.com/paysuper/paysuper-checkout/pkg/micro"
	"html/template"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
)

// Dispatcher
type Dispatcher struct {
	ctx    context.Context
	cfg    Config
	appSet AppSet
	provider.LMT
	globalCfg *common.Config
	ms        *micro.Micro
}

// dispatch
func (d *Dispatcher) Dispatch(echoHttp *echo.Echo) error {
	t, e := template.New("").Funcs(common.FuncMap).ParseGlob(d.cfg.WorkDir + "/assets/web/template/*.html")
	if e != nil {
		return e
	}
	echoHttp.Renderer = common.NewTemplate(t)
	echoHttp.Binder = &common.Binder{}
	// Called after routes
	echoHttp.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Output: logger.NewLevelWriter(d.L(), logger.LevelInfo),
		Format: `{"id":"${id}","remote_ip":"${remote_ip}",` +
			`"host":"${host}","method":"${method}","uri":"${uri}","user_agent":"${user_agent}",` +
			`"status":${status},"error":"${error}","latency":${latency},"latency_human":"${latency_human}"` +
			`,"bytes_in":${bytes_in},"bytes_out":${bytes_out}}`,
	})) // 3

	allowOrigins := strings.Split(d.globalCfg.AllowOrigin, ",")

	echoHttp.Use(d.RecoverMiddleware()) // 3
	echoHttp.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     allowOrigins,
		AllowCredentials: true,
		AllowHeaders:     []string{"authorization", "content-type"},
		ExposeHeaders:    []string{"authorization", "content-type", "set-cookie", "cookie"},
	})) // 2
	// Called before routes
	echoHttp.Use(d.RawBodyPreMiddleware) // 1
	// init group routes
	grp := &common.Groups{
		Common: echoHttp.Group(common.NoAuthGroupPath),
	}
	// init routes
	for _, handler := range d.appSet.Handlers {
		handler.Route(grp)
	}
	if d.cfg.PathRouteDump != "" {
		d.dumpRoutesToFile(echoHttp)
	}
	return nil
}

func (d *Dispatcher) dumpRoutesToFile(echoHttp *echo.Echo) {

	var list []string

	strRepl := strings.NewReplacer("github.com/paysuper/paysuper-checkout/internal/handlers.", "", "-fm", "")
	rts := echoHttp.Routes()

	for _, r := range rts {
		if strings.Contains(r.Name, "v4.glob..func1") {
			continue
		}
		list = append(list, r.Path+" "+r.Method+" "+strRepl.Replace(r.Name))
	}

	sort.Strings(list)

	table := simpletable.New()

	table.Header = &simpletable.Header{
		Cells: []*simpletable.Cell{
			{Align: simpletable.AlignCenter, Text: "Path"},
			{Align: simpletable.AlignCenter, Text: "Method"},
			{Align: simpletable.AlignCenter, Text: "Handler"},
		},
	}

	for _, sl := range list {
		row := strings.Split(sl, " ")
		r := []*simpletable.Cell{
			{Align: simpletable.AlignLeft, Text: row[0]},
			{Align: simpletable.AlignLeft, Text: row[1]},
			{Align: simpletable.AlignLeft, Text: row[2]},
		}
		table.Body.Cells = append(table.Body.Cells, r)
	}

	table.SetStyle(simpletable.StyleUnicode)

	if e := ioutil.WriteFile(d.cfg.PathRouteDump, []byte(table.String()), 0777); e != nil {
		d.L().Error("routes dump can't save to %v, err %v", logger.Args(d.cfg.PathRouteDump, e.Error()))
		return
	}

	d.L().Info("routes dump successfully saved to %v", logger.Args(d.cfg.PathRouteDump))
}

func (d *Dispatcher) commonRoutes(echoHttp *echo.Echo) {
	echoHttp.Static("/", d.cfg.WorkDir+"/assets/web/static")
	echoHttp.Static("/spec", d.cfg.WorkDir+"/api")
	echoHttp.GET("/docs", func(ctx echo.Context) error {
		return ctx.Render(http.StatusOK, "docs.html", map[string]interface{}{})
	})
}

// Config
type Config struct {
	Debug         bool `fallback:"shared.debug"`
	WorkDir       string
	PathRouteDump string
	invoker       *invoker.Invoker
}

// OnReload
func (c *Config) OnReload(callback func(ctx context.Context)) {
	c.invoker.OnReload(callback)
}

// Reload
func (c *Config) Reload(ctx context.Context) {
	c.invoker.Reload(ctx)
}

// AppSet
type AppSet struct {
	Handlers common.Handlers
	Services common.Services
}

// New
func New(ctx context.Context, set provider.AwareSet, appSet AppSet, cfg *Config, globalCfg *common.Config, ms *micro.Micro) *Dispatcher {
	set.Logger = set.Logger.WithFields(logger.Fields{"service": common.Prefix})
	return &Dispatcher{
		ctx:       ctx,
		cfg:       *cfg,
		appSet:    appSet,
		LMT:       &set,
		globalCfg: globalCfg,
		ms:        ms,
	}
}
