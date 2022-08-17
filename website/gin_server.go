package website

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
)

func (s *server) StartServing(context.Context) error {
	router := gin.Default()
	router.LoadHTMLGlob(s.config.WebsitePaths.TemplatesRootDir + "/*")

	router.GET("/", func(c *gin.Context) {
		s.renderIndex(c)
	})

	router.GET("/index", func(c *gin.Context) {
		s.renderIndex(c)
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
	return router.Run(":8080")
}
