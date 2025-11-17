package router

import (
	"embed"
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DashboardRenderer struct {
	templates *template.Template
}

func NewDashboardRenderer(fs embed.FS, templatePath string) (*DashboardRenderer, error) {
	tmpl, err := template.ParseFS(fs, templatePath)
	if err != nil {
		return nil, err
	}

	return &DashboardRenderer{
		templates: tmpl,
	}, nil
}

// Render implements gin's HTML renderer
func (r *DashboardRenderer) Render(c *gin.Context, templateName string, data interface{}) {
	c.Header("Content-Type", "text/html")
	err := r.templates.ExecuteTemplate(c.Writer, templateName, data)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
}
