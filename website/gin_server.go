package website

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/contrib/gzip"
	"github.com/gin-gonic/gin"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"stocks/utils"
	"time"
)

func (s *server) StartServing(ctx context.Context, kill time.Duration) error {
	router := gin.New()

	router.Use(gzip.Gzip(gzip.DefaultCompression))

	s.setupPanicAndFailureHandlers(router)

	s.setupRobotsAndSiteMap(router)

	infos, err := ioutil.ReadDir(themePath)
	utils.PanicErr(err)

	for _, info := range infos {
		if info.IsDir() {
			router.Static(fmt.Sprintf("/%s", info.Name()), fmt.Sprintf("%s/%s", themePath, info.Name()))
		}
	}
	router.Static("/static", staticPath)
	router.SetFuncMap(template.FuncMap{
		"renderETFsArray":         renderETFsArray,
		"renderPercentage":        renderPercentage,
		"renderLargeNumbers":      renderLargeNumbers,
		"renderStockTickersCount": renderStockTickersCount,
		"renderDate":              renderDate,
	})
	router.LoadHTMLGlob(s.metadata.TemplateCustomMetadata.WebsitePaths.TemplatesRootDir + "/**/*")

	var index = []string{"", "index", "index.html", "find_overlaps.html", "find_overlaps"}

	s.route(index, router, func(c *gin.Context) {
		s.renderFindOverlapsInputHTML(c)
	})

	router.GET("/etf-summary/overlap", func(c *gin.Context) {
		s.renderOverlap(c)
	})

	router.GET("/etf-summary/:etf", func(c *gin.Context) {
		s.renderETF(c)
	})

	router.GET(fmt.Sprintf("/stock-summary/:%s", stockParamKey), func(c *gin.Context) {
		s.renderStock(c)
	})

	router.GET("/use_cases.html", func(c *gin.Context) {
		s.renderUseCases(c)
	})

	router.GET("/disclaimer.html", func(c *gin.Context) {
		s.renderDisclaimer(c)
	})

	router.POST("/find_overlaps.html", func(c *gin.Context) {
		s.findOverlapsForCustomHoldings(c)
	})

	router.GET("/list_all_etfs.html", func(c *gin.Context) {
		s.renderAllETFs(c)
	})

	router.GET("/list_all_stocks.html", func(c *gin.Context) {
		s.renderAllStocks(c)
	})

	router.GET("/faq.html", func(c *gin.Context) {
		s.renderFAQs(c)
	})

	router.GET("/contact.html", func(c *gin.Context) {
		s.renderContactPage(c)
	})

	if kill > time.Second {
		fmt.Printf("Configured to be killed in %v\n", kill)
		s.setupToKill(ctx, kill, router)
	} else {
		return router.Run(addr)
	}
	return nil
}

func (s *server) setupPanicAndFailureHandlers(router *gin.Engine) {
	router.Use(gin.CustomRecovery(func(c *gin.Context, err interface{}) {
		c.HTML(http.StatusInternalServerError, error404tmpl, s.commonStruct())
	}))

	router.NoRoute(func(c *gin.Context) {
		c.HTML(http.StatusInternalServerError, error404tmpl, s.commonStruct())
	})
	router.NoMethod(func(c *gin.Context) {
		c.HTML(http.StatusInternalServerError, error404tmpl, s.commonStruct())
	})
}

func (s *server) route(paths []string, router *gin.Engine, handler func(c *gin.Context)) {
	for _, path := range paths {
		router.GET(path, handler)
	}
}
func (s *server) setupRobotsAndSiteMap(r *gin.Engine) {
	r.StaticFile("robots.txt", fmt.Sprintf("%s/robots.txt", generatedPath))
	r.StaticFile("sitemap.xml", fmt.Sprintf("%s/sitemap.xml", generatedPath))
}

const addr = ":8080"

func (s *server) setupToKill(ctx context.Context, kill time.Duration, engine *gin.Engine) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	srv := &http.Server{
		Addr:    addr,
		Handler: engine,
	}

	// Initializing the server in a goroutine so that
	// it won't block the graceful shutdown handling below
	go func() {
		if err := srv.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
			log.Printf("listen: %s\n", err)
		}
	}()

	time.Sleep(kill)

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}
