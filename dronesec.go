package dronesec

import (
	"errors"
	"fmt"
	"strings"

	"github.com/alexander-localbitcoins/dronesec/client"
	"github.com/alexander-localbitcoins/logger"

	"github.com/drone/drone-go/drone"
)

type DroneSecOptions uint32

const (
	Insecure DroneSecOptions = 1 << iota // Ignore invalid certificates
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
	if _, err := d.client.Secret(d.owner, d.repo, secret); err != nil && strings.Contains(err.Error(), "no rows in result set") {
		_, err := d.client.SecretCreate(d.owner, d.repo, sec)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else {
		d.log.Warning("Overwriting old secret")
		if _, err := d.client.SecretUpdate(d.owner, d.repo, sec); err != nil {
			return err
		}
	}
	return nil
}
