package enforcer

import (
	"fmt"

	"code.cloudfoundry.org/lager"
	"quota-enforcer/database"
)

type Enforcer interface {
	EnforceOnce() error
}

type enforcer struct {
	violatorRepo, reformerRepo database.Repo
	logger                     lager.Logger
}

func NewEnforcer(violatorRepo, reformerRepo database.Repo, logger lager.Logger) Enforcer {
	return &enforcer{
		violatorRepo: violatorRepo,
		reformerRepo: reformerRepo,
		logger:       logger,
	}
}

func (e enforcer) EnforceOnce() error {
	err := e.revokePrivilegesFromViolators()
	if err != nil {
		return err
	}

	err = e.grantPrivilegesToReformed()
	if err != nil {
		return err
	}

	return nil
}

func (e enforcer) revokePrivilegesFromViolators() error {
	e.logger.Info("Looking for violators")

	violators, err := e.violatorRepo.All()
	if err != nil {
		return fmt.Errorf("Finding violators: %s", err.Error())
	}

	for _, db := range violators {
		err = db.RevokePrivileges()
		if err != nil {
			return fmt.Errorf("Revoking privileges: %s", err.Error())
		}

		err = db.KillActiveConnections()
		if err != nil {
			return fmt.Errorf("Resetting active privileges: %s", err.Error())
		}
	}
	return nil
}

func (e enforcer) grantPrivilegesToReformed() error {
	e.logger.Info("Looking for reformers")

	reformers, err := e.reformerRepo.All()
	if err != nil {
		return fmt.Errorf("Finding reformers: %s", err.Error())
	}

	for _, db := range reformers {
		err = db.GrantPrivileges()
		if err != nil {
			return fmt.Errorf("Granting privileges: %s", err.Error())
		}

		err = db.KillActiveConnections()
		if err != nil {
			return fmt.Errorf("Resetting active privileges: %s", err.Error())
		}
	}

	return nil
}
