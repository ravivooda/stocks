package website

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"net/http"
	"stocks/utils"
	"time"
)

func (s *server) StartServing(ctx context.Context, kill time.Duration) error {
	router := gin.Default()

	dirname := "./website/letf/static/focus/theme"
	infos, err := ioutil.ReadDir(dirname)
	utils.PanicErr(err)

	for _, info := range infos {
		if info.IsDir() {
			router.Static(fmt.Sprintf("/%s", info.Name()), fmt.Sprintf("%s/%s", dirname, info.Name()))
		}
	}
	router.LoadHTMLGlob(s.metadata.TemplateCustomMetadata.WebsitePaths.TemplatesRootDir + "/**/*")

	router.GET("/", func(c *gin.Context) {
		s.renderAllETFs(c)
	})

	router.GET("/index", func(c *gin.Context) {
		s.renderAllETFs(c)
	})

	router.GET(fmt.Sprintf("/etf-summary/overlap/:%s", overlapParam), func(c *gin.Context) {
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

	router.GET("/find_overlaps.html", func(c *gin.Context) {
		s.renderFindOverlapsInputHTML(c)
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

	if kill > time.Second {
		fmt.Printf("Configured to be killed in %v\n", kill)
		s.setupToKill(ctx, kill, router)
	} else {
		return router.Run(addr)
	}
	return nil
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
