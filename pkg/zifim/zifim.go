package zifim

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/dghubble/sling"
	"github.com/labstack/echo/v4"
	"github.com/phuslu/log"
	"github.com/samber/lo"
	"github.com/shmul/bekind/pkg/web"
)

const (
	githubTreesURL = "https://api.github.com/repos/shmul/zifim/git/trees/HEAD?recursive=true"
)

type Module struct {
	shows []string
	l     log.Logger
}

type githubNode struct {
	Mode string `json:"mode"`
	Path string `json:"path"`
	SHA  string `json:"sha"`
	Size int64  `json:"size"`
	Type string `json:"type"`
	URL  string `json:"url"`
}
type githubTreeContent struct {
	SHA       string       `json:"sha"`
	Tree      []githubNode `json:"tree"`
	Truncated bool         `json:"truncated"`
	URL       string       `json:"url"`
}

func New() *Module {
	m := &Module{
		l: log.DefaultLogger,
	}

	go func() {
		var helper func()
		helper = func() {
			newShows := m.getDirectoryContent()
			if len(newShows) > len(m.shows) {
				m.shows = newShows
			}
			time.AfterFunc(10*time.Minute, helper)
		}
		helper()
	}()

	return m
}

func (z *Module) Setup(wb *web.Web, g *echo.Group) {
	g.GET("/show/:date", z.show)
	g.GET("/shows", func(c echo.Context) error {
		return c.JSON(http.StatusOK, z.shows)
	})
}

func (z *Module) show(c echo.Context) error {
	d := c.Param("date") // format yy-mm-dd
	parts := strings.Split(d, "-")
	if len(parts) != 3 {
		z.l.Warn().Str("date", d).Msg("show")
		return echo.NewHTTPError(http.StatusBadRequest, d)
	}
	githubPath := fmt.Sprintf("https://raw.githubusercontent.com/shmul/zifim/master/20%s/%s.txt",
		parts[0], d)

	resp, err := http.Get(githubPath)
	if err != nil {
		z.l.Warn().Err(err).Msg("show")
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	defer resp.Body.Close()
	out, err := io.ReadAll(resp.Body)
	if err != nil {
		z.l.Warn().Err(err).Msg("show")
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	lines := strings.FieldsFunc(string(out), func(c rune) bool { return c == '\n' || c == '\r' })

	return c.String(http.StatusOK, strings.Join(lines, "<br>"))
}

func (z *Module) getDirectoryContent() []string {
	var tree githubTreeContent
	_, err := sling.New().Get(githubTreesURL).ReceiveSuccess(&tree)
	if err != nil {
		return nil
	}
	shows := lo.Map(
		lo.Filter(tree.Tree, func(n githubNode, _ int) bool {
			return n.Type == "blob" && strings.HasPrefix(n.Path, "20") && strings.HasSuffix(n.Path, ".txt")
		}), func(n githubNode, _ int) string {
			s := strings.Index(n.Path, "/")
			e := strings.Index(n.Path, ".")
			return n.Path[s+1 : e]
		})
	z.l.Info().Int("shows", len(shows)).Msg("getDirectoryContent")
	return shows
}
