package command

import (
	"strings"
	"testing"

	"go.mlcdf.fr/owh/internal/api"
	"go.mlcdf.fr/owh/internal/config"
	"go.mlcdf.fr/owh/internal/view"
	"go.mlcdf.fr/owh/tests/smockertest"
)

func TestWhoami(t *testing.T) {
	smockertest.MustStart(t)
	defer smockertest.Container.Nuke(t)

	buf := &strings.Builder{}

	app := App{
		APIClientFactory: api.NewClient,
		HTTPClient:       smockertest.HTTPClient,
		View:             &view.View{Writer: buf},
		Config: &config.Config{
			Region:      "ovh-eu",
			ConsumerKey: "yolo",
		},
	}

	err := smockertest.PushMock("../../tests/mocks/whoami.yaml")
	if err != nil {
		t.Error(err)
	}

	cmd := WhoamiCommand{App: app}
	exitCode := cmd.Run([]string{})

	if exitCode != 0 {
		t.Errorf("exit code, want 0, got %d", exitCode)
		t.Logf("%s", buf)
	}

	want := `  Name = Maxime Le Conte des Floris
  ID   = xxxxxxx
`

	if got := buf.String(); got != want {
		t.Errorf("want '%s', got '%s'", want, got)
	}
}
