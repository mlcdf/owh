package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path"

	_ "embed"

	"github.com/adrg/xdg"
	"go.mlcdf.fr/owh/internal/cmdutil"
	"go.mlcdf.fr/sally/logging"
	"golang.org/x/xerrors"
)

// GlobalOpts holds the global owh configuration
var GlobalOpts *globalOptions

type globalOptions struct {
	// config file location on disk
	location string `json:"-"`

	Region          string                  `json:"region,omitempty"`
	ConsumerKey     string                  `json:"consumer_key,omitempty"`
	SFTPCredentials map[string]*Credentials `json:"ssh_passwords,omitempty"`
}

type Credentials struct {
	User     string `json:"user"`
	Password string `json:"password"`
}

func New() error {
	location, err := xdg.ConfigFile("owh/config.json")
	if err != nil {
		return err
	}

	GlobalOpts = &globalOptions{location: location}

	err = fromFile(GlobalOpts, location)
	if err != nil {
		return err
	}

	if GlobalOpts.SFTPCredentials == nil {
		GlobalOpts.SFTPCredentials = map[string]*Credentials{}
	}

	fromEnv(GlobalOpts)
	return nil
}

func (opts *globalOptions) Validate() error {
	if opts.Region == "" || opts.ConsumerKey == "" {
		if ci := os.Getenv("CI"); ci != "" {
			fmt.Println("To use owh in automation, set the OWH_CONSUMER_KEY environment variable.")
		} else {
			fmt.Println("Please run: owh login first.\nAlternatively, populate the OWH_CONSUMER_KEY environment variable with your consumer key.")
		}
		return cmdutil.ErrSilent
	}

	if opts.Region != "ovh-eu" && opts.Region != "ovh-ca" {
		fmt.Printf("Invalid owh configuration file format: %s. Try to remove it and run: owh login", cmdutil.Color(cmdutil.StyleHighlight).Render(opts.location))
		return cmdutil.ErrSilent
	}

	return nil
}

func fromEnv(opts *globalOptions) {
	if consumerKey := os.Getenv("OWH_CONSUMER_KEY"); consumerKey != "" {
		opts.ConsumerKey = consumerKey
	}

	region := os.Getenv("OWH_REGION")

	if region != "" {
		opts.Region = region
	} else if region == "" && !cmdutil.IsInteractive() {
		opts.Region = "ovh-eu"
		logging.Debugf("OWH_REGION environment variable not set. Defaulting to '%s'.", opts.Region)
	}
}

func fromFile[Options *globalOptions | *Link](opts Options, location string) error {
	if _, err := os.Stat(location); os.IsNotExist(err) {
		logging.Debugf("Config file %s does not exist", location)
	} else {
		logging.Debugf("Config file found at %s", location)
		fh, err := os.Open(location)
		if err != nil {
			return err
		}
		defer fh.Close()

		decoder := json.NewDecoder(fh)
		if err := decoder.Decode(&opts); err != nil {
			fmt.Printf("Folder link is invalid %s. Try to remove it and run: owh login", cmdutil.Color(cmdutil.StyleHighlight).Render(location))
			return cmdutil.ErrSilent
		}

	}

	logging.Debugf("%v", opts)
	return nil
}

func Save[Options *globalOptions | *Link](opts Options) error {
	var location string
	// begin ugly
	// I'm waiting for https://github.com/golang/go/issues/45380
	if o, ok := any(opts).(*globalOptions); ok {
		location = o.location
	}

	if o, ok := any(opts).(*Link); ok {
		location = o.location
	}
	// end ugly

	if location == "" {
		return xerrors.Errorf("what the fuck")
	}

	var fh *os.File

	if _, err := os.Stat(location); os.IsNotExist(err) {
		if err := os.MkdirAll(path.Dir(location), 0600); err != nil {
			return err
		}

		fh, err = os.Create(location)
		if err != nil {
			return err
		}
	} else {
		fh, err = os.OpenFile(location, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			return err
		}
	}

	defer fh.Close()

	bytes, err := json.MarshalIndent(&opts, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode json: %w", err)
	}

	_, err = fh.Write(bytes)
	if err != nil {
		return fmt.Errorf("failed to write to file %s: %w", location, err)
	}

	logging.Infof("config saved")

	return nil
}
