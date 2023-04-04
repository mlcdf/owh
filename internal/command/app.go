package command

import (
	"net/http"

	"go.mlcdf.fr/owh/internal/api"
	"go.mlcdf.fr/owh/internal/config"
	"go.mlcdf.fr/owh/internal/view"
)

// App holds dependencies used by most commands.
type App struct {
	APIClientFactory api.ClientFactory
	Config           *config.Config
	HTTPClient       *http.Client
	IsInteractive    bool
	LinkFunc         config.LinkFactory
	View             *view.View
}

func (app *App) EnsureLink() (*config.Link, error) {
	return app.LinkFunc(app.IsInteractive)
}

func (app *App) LoggedClient() (*api.Client, error) {
	if err := app.Config.IsValid(); err != nil {
		return nil, err
	}

	return app.APIClientFactory(app.HTTPClient, app.Config.Region, app.Config.ConsumerKey)
}
