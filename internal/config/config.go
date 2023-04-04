package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/adrg/xdg"
	"go.mlcdf.fr/owh/internal/cmdutil"
	"go.mlcdf.fr/sally/logging"
	"golang.org/x/xerrors"
)

const ENV_PREFIX = "OWH_"
const ENV_REGION = ENV_PREFIX + "REGION"
const ENV_CONSUMER_KEY = ENV_PREFIX + "CONSUMER_KEY"
const ENV_SSH_USER = ENV_PREFIX + "SSH_USER"
const ENV_SSH_PASSWORD = ENV_PREFIX + "SSH_PASSWORD"

type Factory func(isInteractive bool) (*Config, error)

type Config struct {
	location string `json:"-"`

	Region          string                  `json:"region,omitempty"`
	ConsumerKey     string                  `json:"consumer_key,omitempty"`
	SFTPCredentials map[string]*Credentials `json:"ssh_passwords,omitempty"`
}

type Credentials struct {
	User     string `json:"user"`
	Password string `json:"password"`
}

var _ Factory = New

func New(isInteractive bool) (*Config, error) {
	location, err := xdg.ConfigFile("owh/config.json")
	if err != nil {
		return nil, err
	}

	config := &Config{location: location}

	err = fromFile(config, location)
	if err != nil {
		return nil, err
	}

	if config.SFTPCredentials == nil {
		config.SFTPCredentials = map[string]*Credentials{}
	}

	fromEnv(config, isInteractive)
	return config, nil
}

func (config *Config) IsValid() error {
	if config.Region == "" || config.ConsumerKey == "" {
		if ci := os.Getenv("CI"); ci != "" {
			fmt.Printf("To use owh in automation, set the %s environment variable.\n", ENV_CONSUMER_KEY)
		} else {
			fmt.Printf(
				"Please run: owh login first.\nAlternatively, populate the %s environment variable with your consumer key.\n",
				ENV_CONSUMER_KEY,
			)
		}
		return cmdutil.ErrSilent
	}

	if config.Region != "ovh-eu" && config.Region != "ovh-ca" {
		fmt.Printf(
			"Invalid owh configuration file format: %s. Try to remove it and run: owh login",
			cmdutil.Color(cmdutil.StyleHighlight).Render(config.location),
		)
		return cmdutil.ErrSilent
	}

	return nil
}

func fromEnv(config *Config, isInteractive bool) {
	if consumerKey := os.Getenv(ENV_CONSUMER_KEY); consumerKey != "" {
		config.ConsumerKey = consumerKey
	}

	region := os.Getenv(ENV_REGION)

	if region != "" {
		config.Region = region
	} else if region == "" && !isInteractive {
		config.Region = "ovh-eu"
		logging.Debugf("%s environment variable not set. Defaulting to '%s'.", ENV_REGION, config.Region)
	}
}

func fromFile[Options *Config | *Link](opts Options, location string) error {
	if _, err := os.Stat(location); os.IsNotExist(err) {
		logging.Debugf("Config file %s does not exist", location)
		return nil
	}
	logging.Debugf("Config file found at %s", location)

	fh, err := os.Open(location)
	if err != nil {
		return err
	}

	defer fh.Close()

	decoder := json.NewDecoder(fh)
	if err := decoder.Decode(&opts); err != nil {
		fmt.Printf(
			"Folder link is invalid %s. Try to remove it and run: owh login",
			cmdutil.Color(cmdutil.StyleHighlight).Render(location),
		)
		return cmdutil.ErrSilent
	}
	return nil
}

func (config *Config) Save() error {
	if config.location == "" {
		// in testing
		return nil
	}

	return save(config)
}

func save[Options *Config | *Link](opts Options) error {
	var location string
	// begin ugly
	// I'm waiting for https://github.com/golang/go/issues/45380
	if o, ok := any(opts).(*Config); ok {
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
