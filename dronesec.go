package dronesec

import (
	"errors"
	"fmt"
	"strings"

	"github.com/alexander-localbitcoins/dronesec/client"
	"github.com/alexander-localbitcoins/logger"

	"github.com/drone/drone-go/drone"
)

var EmptyInput = errors.New("Input cannot be empty")

// Struct for creating drone secrets. If log is nil logging will be disabled
func NewDroneSec(host string, owner string, repo string, token string, certsLocation string, flags client.BuilderFlags, log logger.Logger) (*droneSec, error) {
	d := new(droneSec)
	switch {
	case host == "":
		return nil, fmt.Errorf("%w: host url", EmptyInput)
	case owner == "":
		return nil, fmt.Errorf("%w: repo owner", EmptyInput)
	case repo == "":
		return nil, fmt.Errorf("%w: repo name", EmptyInput)
	case token == "":
		return nil, fmt.Errorf("%w: token", EmptyInput)
	}
	d.owner = owner
	d.repo = repo
	if log == nil {
		log = logger.NewLogger(logger.Empty)
	}
	d.log = log
	db, err := client.NewDroneBuilder(log, certsLocation, flags)
	if err != nil {
		return nil, err
	}
	d.client, err = db.BuildDroneClient(host, token)
	if err != nil {
		return nil, err
	}
	return d, nil
}

type droneSec struct {
	token  string
	repo   string
	owner  string
	client drone.Client
	log    logger.Logger
}

func (d *droneSec) Create(secret string, data string) error {
	sec := &drone.Secret{
		Name:        secret,
		Data:        data,
		PullRequest: true,
	}
	if _, err := d.client.SecretCreate(d.owner, d.repo, sec); err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed: secrets.secret_repo_id, secrets.secret_name") {
			d.log.Debug(err)
			d.log.Warning("Overwriting old secret")
			if _, err := d.client.SecretUpdate(d.owner, d.repo, sec); err != nil {
				return err
			}
		} else {
			return err
		}
	}
	d.log.Info(fmt.Sprintf("Succesfully pushed %s to drone server", secret))
	return nil
}

func (d *droneSec) Delete(secret string) error {
	if err := d.client.SecretDelete(d.owner, d.repo, secret); err != nil {
		if strings.Contains(err.Error(), "sql: no rows in result set") {
			d.log.Debug(err)
			d.log.Warning("Secret not found, could not delete")
			return nil
		}
		return err
	}
	d.log.Info(fmt.Sprintf("Deleted secret %s", secret))
	return nil
}
