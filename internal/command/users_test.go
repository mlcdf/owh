package command

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mlcdf.fr/owh/internal/api"
	"go.mlcdf.fr/owh/internal/config"
	"go.mlcdf.fr/owh/internal/view"
	"go.mlcdf.fr/owh/tests/smockertest"
)

func TestUsers(t *testing.T) {
	smockertest.MustStart(t)
	defer smockertest.Container.Nuke(t)

	buf := &strings.Builder{}

	app := App{
		LinkFunc: func(isInteractive bool) (*config.Link, error) {
			return &config.Link{Hosting: "asterix.cluster031.hosting.ovh.net", CanonicalDomain: "yolo.fr"}, nil
		},
		APIClientFactory: api.NewClient,
		HTTPClient:       smockertest.HTTPClient,
		View:             &view.View{Writer: buf},
		Config: &config.Config{
			Region:      "ovh-eu",
			ConsumerKey: "yolo",
		},
	}

	err := smockertest.PushMock("../../tests/mocks/users.yaml")
	if err != nil {
		t.Error(err)
	}

	cmd := UsersCommand{App: app}
	exitCode := cmd.Run([]string{})

	if exitCode != 0 {
		t.Errorf("exit code, want 0, got %d", exitCode)
		t.Logf("%s", buf)
	}

	want := `LOGIN      	PRIMARY LOGIN 
asterix    	true         	
asterix-owh	false        	
`
	assert.Equal(t, want, buf.String())
}
