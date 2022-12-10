package flow

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
	"unicode"

	"github.com/AlecAivazis/survey/v2"
	"go.mlcdf.fr/owh/internal/api"
	"go.mlcdf.fr/owh/internal/cmdutil"
	"go.mlcdf.fr/owh/internal/view"

	cfg "go.mlcdf.fr/owh/internal/config"
	"go.mlcdf.fr/owh/internal/remote"
	"golang.org/x/xerrors"
)

func NewSSHClient(client *api.Client, config *cfg.Config, view *view.View, isInteractive bool, hosting string) (*remote.Client, error) {
	hostingInfo, err := client.GetHosting(hosting)
	if err != nil {
		return nil, err
	}

	credentials, ok := config.SFTPCredentials[hosting]
	if !ok {
		credentials = &cfg.Credentials{}

		if password := os.Getenv(cfg.ENV_SSH_PASSWORD); password != "" {
			credentials.Password = password
		}

		if user := os.Getenv(cfg.ENV_SSH_USER); user != "" {
			credentials.User = user
		} else {
			credentials.User = hostingInfo.PrimaryLogin
		}
	}

	if credentials.Password == "" {
		if !isInteractive {
			fmt.Println("No SSH credentials found in config or environnement variable.")
			return nil, cmdutil.ErrSilent
		}

		var input string

		OPTION_CREATE_NEW_USER := "Create a new ssh user"
		OPTION_RESET_PASSWORD := "Reset the password of the existing ssh user"
		OPTION_ABORT_DEPLOYMENT := "Abort the deployment"

		prompt := &survey.Select{
			Message: "No SSH credentials found in config or environnement variable. What would you like to do?",
			Options: []string{OPTION_CREATE_NEW_USER, OPTION_RESET_PASSWORD, OPTION_ABORT_DEPLOYMENT},
		}
		err := survey.AskOne(prompt, &input)
		if err != nil {
			return nil, xerrors.Errorf("failed to display prompt %w", err)
		}

		switch input {
		case OPTION_CREATE_NEW_USER:
			credentials, err = createSSHUser(client, view, config, hostingInfo.PrimaryLogin, hosting)
			if err != nil {
				return nil, err
			}
			fmt.Printf("SSH user %s created\n", credentials.User)
		case OPTION_RESET_PASSWORD:
			users, err := client.ListUsers(hosting)
			if err != nil {
				return nil, err
			}

			var user string
			prompt := &survey.Select{Message: "Users", Options: users}

			err = survey.AskOne(prompt, &user)
			if err != nil {
				return nil, err
			}

			err = ChangePassword(client, config, hosting, user, "")
			if err != nil {
				return nil, err
			}
		case OPTION_ABORT_DEPLOYMENT:
			fallthrough
		default:
			fmt.Println("Aborting the deployement")
			return nil, cmdutil.ErrCancel
		}
	}

	sshConfig := remote.NewPasswordConfig(
		hostingInfo.ServiceManagementAccess.SSH.URL,
		hostingInfo.ServiceManagementAccess.SSH.Port,
		credentials.User,
		credentials.Password,
	)

	conn, err := remote.Connect(sshConfig)

	if err != nil {
		return nil, err
	}

	return conn, nil
}

func createSSHUser(client *api.Client, view *view.View, config *cfg.Config, primaryLogin string, hosting string) (*cfg.Credentials, error) {
	login := fmt.Sprintf("%s-owh", primaryLogin)
	prompt := &survey.Input{
		Message: "SSH user",
		Default: login,
		Help:    fmt.Sprintf("It should start with %s-", primaryLogin),
		// TODO: validate input
	}

	err := survey.AskOne(prompt, &login, survey.WithValidator(survey.Required))
	if err != nil {
		return nil, xerrors.Errorf("failed to display prompt %w", err)
	}

	password := GenPassword()
	payload := &api.SSHUser{
		Home:     ".",
		Login:    login,
		Password: password,
		SSHState: "active",
	}

	var task *api.Task
	url := fmt.Sprintf("/hosting/web/%s/user", hosting)
	err = client.Post(url, &payload, &task)
	if err != nil {
		return nil, xerrors.Errorf("failed to create SSH user: %w", err)
	}

	err = WaitTaskDone(client, view, hosting, task.ID, fmt.Sprintf("Creating SSH user %s", login))
	if err != nil {
		var response interface{}
		err := client.Get(url, response)
		if err != nil {
			return nil, xerrors.Errorf("failed to create SSH user: %w", err)
		}
	}

	credentials := &cfg.Credentials{User: login, Password: password}
	config.SFTPCredentials[hosting] = credentials

	return credentials, config.Save()
}

func GenPassword() string {
	rand.Seed(time.Now().UnixNano())
	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		"0123456789")
	length := 19

	var b strings.Builder
	for i := 0; i < length; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}

	return b.String()
}

func validatePassword(password string) bool {
	letters := 0

	for _, c := range password {
		switch {
		case unicode.IsPunct(c) || unicode.IsSymbol(c):
			return false
		case unicode.IsLetter(c) || c == ' ':
			letters++
		default:
		}
	}

	if letters < 8 || letters > 20 {
		return false
	}

	return true
}

func ChangePassword(client *api.Client, conf *cfg.Config, hosting string, user string, password string) error {
	if password == "" {
		prompt := &survey.Input{
			Message: "Password (alphanumeric characters only, leave blank to use an auto-generated password)",
		}

		err := survey.AskOne(
			prompt,
			&password,
			survey.WithValidator(survey.MaxLength(20)),
		)

		if err != nil {
			return err
		}

		if password == "" {
			password = GenPassword()
		}

		if !validatePassword(password) {
			fmt.Printf("Invalid password format %s\n", cmdutil.Color(cmdutil.StyleHighlight).Render(password))
			return cmdutil.ErrSilent
		}
	}

	_, err := client.ChangePassword(hosting, user, password)
	if err != nil {
		return err
	}

	conf.SFTPCredentials[hosting] = &cfg.Credentials{User: user, Password: password}
	err = conf.Save()
	if err != nil {
		return err
	}

	fmt.Printf("Password for user %s changed\n", cmdutil.Color(cmdutil.StyleHighlight).Render(user))
	return nil
}
