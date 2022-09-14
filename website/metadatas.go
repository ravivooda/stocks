package website

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func (s *server) renderUseCases(c *gin.Context) {
	c.HTML(http.StatusOK, useCasesTemplate, nil)
}

func (s *server) renderDisclaimer(c *gin.Context) {
	c.HTML(http.StatusOK, disclaimerTemplate, nil)
}
