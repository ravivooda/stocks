package website

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func (s *server) renderUseCases(c *gin.Context) {
	c.HTML(http.StatusOK, useCasesTemplate, s.commonStruct())
}

func (s *server) renderDisclaimer(c *gin.Context) {
	c.HTML(http.StatusOK, disclaimerTemplate, s.commonStruct())
}

func (s *server) renderFAQs(c *gin.Context) {
	c.HTML(http.StatusOK, faqsTemplate, s.commonStruct())
}

func (s *server) renderContactPage(c *gin.Context) {
	c.HTML(http.StatusOK, contactTemplate, s.commonStruct())
}

func (s *server) commonStruct() struct{ TemplateCustomMetadata TemplateCustomMetadata } {
	return struct {
		TemplateCustomMetadata TemplateCustomMetadata
	}{
		TemplateCustomMetadata: s.metadata.TemplateCustomMetadata,
	}
}
