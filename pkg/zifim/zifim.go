package zifim

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/shmul/bekind/pkg/web"
)

type Module struct{}

func (z *Module) Setup(wb *web.Web, g *echo.Group) {
	g.GET("/show/:date", z.show)
}

func (z *Module) show(c echo.Context) error {
	d := c.Param("date") // format yy-mm-dd
	parts := strings.Split(d, "-")
	if len(parts) != 3 {
		return echo.NewHTTPError(http.StatusBadRequest, d)
	}
	githubPath := fmt.Sprintf("https://raw.githubusercontent.com/shmul/zifim/master/20%s/%s.txt",
		parts[0], d)
	resp, err := http.Get(githubPath)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	defer resp.Body.Close()
	out, err := io.ReadAll(resp.Body)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	lines := strings.FieldsFunc(string(out), func(c rune) bool { return c == '\n' || c == '\r' })

	return c.String(http.StatusOK, strings.Join(lines, "<br>"))
}
