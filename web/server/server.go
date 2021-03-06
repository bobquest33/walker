package server

import (
	"fmt"
	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
	"github.com/lcaballero/gel"
	"github.com/lcaballero/hitman"
	"github.com/lcaballero/reverb/base"
	"github.com/lcaballero/walker/conf"
	"github.com/lcaballero/walker/searching"
	app "github.com/lcaballero/walker/web/context"
	"log"
	"strconv"
)

type WebServer struct {
	*conf.Config
	searcher         *searching.Searcher
	IncludesResolver func(string) gel.View
}

func NewWebServer(searcher *searching.Searcher, config *conf.Config) (*WebServer, error) {
	ws := &WebServer{
		Config:   config,
		searcher: searcher,
	}
	return ws, nil
}

func (w *WebServer) Start() hitman.KillChannel {
	done := hitman.NewKillChannel()
	go func() {
		go w.run()
		for {
			select {
			case cleaner := <-done:
				cleaner.WaitGroup.Done()
				return
			}
		}
	}()
	return done
}

func (w *WebServer) run() {
	log.Printf("finding templates at: %s", w.IncludesDir)
	log.Printf("using assets found at: %s", w.AssetsDir)

	ctx := app.NewContext(w.Config)

	r := base.NewRegister()
	r.Get("/asset/:kind/:hash/:file", ctx.ToHandler())
	r.Get("/searching", Index(ctx, *w.Config, w.searcher))

	log.Printf("starting web server at: %s", w.Ip)
	log.Println(w.Config.String())
	r.Echo.Run(standard.New(w.Ip))
}

func Index(ctx *app.Context, cfg conf.Config, searcher *searching.Searcher) echo.HandlerFunc {
	return func(c echo.Context) error {
		cfg.Query = c.QueryParam("q")
		cfg.ShowFontLocks = c.QueryParam("show-fonts") != ""
		cfg.ShowQuery = c.QueryParam("show-query") != ""

		max := c.QueryParam("max-results")
		if max != "" {
			n, err := strconv.Atoi(max)
			if err == nil {
				cfg.MaxHits = n
			}
		}

		fmt.Printf("searching for: %s\n", cfg.Query)
		fmt.Printf("max-hits for: %d\n", cfg.MaxHits)

		sr, err := searcher.Query(cfg)
		if err != nil {
			return err
		}

		_, err = c.Response().Write(sr.Render().Bytes())
		return err
	}
}
