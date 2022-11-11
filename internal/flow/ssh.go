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
	"go.mlcdf.fr/owh/internal/config"
	"go.mlcdf.fr/sally/logging"
	"golang.org/x/crypto/ssh"
	"golang.org/x/xerrors"
)

func NewSSHClient(client *api.Client, hosting string) (*ssh.Client, error) {
	hostingInfo, err := client.HostingInfo(hosting)
	if err != nil {
		return nil, err
	}

	credentials, ok := config.GlobalOpts.SFTPCredentials[hosting]
	if !ok {
		credentials = &config.Credentials{}

		if password := os.Getenv(config.ENV_SSH_PASSWORD); password != "" {
			credentials.Password = password
		}

		if user := os.Getenv(config.ENV_SSH_USER); user != "" {
			credentials.User = user
		} else {
			credentials.User = hostingInfo.PrimaryLogin
		}
	}

	if credentials.Password == "" {
		if !cmdutil.IsInteractive() {
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
			credentials, err = createSSHUser(client, hostingInfo.PrimaryLogin, hosting)
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

			err = ChangePassword(client, hosting, user, "")
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

	config := ssh.ClientConfig{
		Config:          ssh.Config{},
		User:            credentials.User,
		Auth:            []ssh.AuthMethod{ssh.Password(credentials.Password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", hostingInfo.ServiceManagementAccess.SSH.URL, hostingInfo.ServiceManagementAccess.SSH.Port)
	logging.Debugf("ssh %s -l %s (passwd: %s)", strings.ReplaceAll(addr, ":22", ""), credentials.User, credentials.Password)

	var conn *ssh.Client

	conn, err = ssh.Dial("tcp", addr, &config)
	if err != nil {
		time.Sleep(5 * time.Second)

		conn, err = ssh.Dial("tcp", addr, &config)
		if err != nil {
			return nil, xerrors.Errorf("failed to connect to [%s]: %w\n", addr, err)
		}
	}

	return conn, nil
}

func createSSHUser(client *api.Client, primaryLogin string, hosting string) (*config.Credentials, error) {
	login := fmt.Sprintf("%s-owh", primaryLogin)
	prompt := &survey.Input{Message: "SSH user", Default: login, Help: fmt.Sprintf("It should start with %s-", primaryLogin)}

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

	err = client.WaitTaskDone(hosting, task.ID)
	if err != nil {
		var response interface{}
		err := client.Get(url, response)
		if err != nil {
			return nil, xerrors.Errorf("failed to create SSH user: %w", err)
		}
	}

	credentials := &config.Credentials{User: login, Password: password}
	config.GlobalOpts.SFTPCredentials[hosting] = credentials

	return credentials, config.GlobalOpts.Save()
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

	if letters < 8 && letters > 20 {
		return false
	}

	return true
}

func ChangePassword(client *api.Client, hosting string, user string, password string) error {
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

	err := client.ChangePassword(hosting, user, password)
	if err != nil {
		return err
	}

	config.GlobalOpts.SFTPCredentials[hosting] = &config.Credentials{User: user, Password: password}
	err = config.GlobalOpts.Save()
	if err != nil {
		return err
	}

	fmt.Printf("Password for user %s changed\n", cmdutil.Color(cmdutil.StyleHighlight).Render(user))
	return nil
}
