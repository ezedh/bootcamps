package secrets

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/google/go-github/v43/github"
)

type SecretManager interface {
	SetSecret(context.Context, string, string) error
}

type secretsManager struct {
	c     *github.Client
	repo  string
	owner string
}

func NewSecretsManager(client *github.Client, owner, repo string) SecretManager {
	return &secretsManager{
		c:     client,
		repo:  repo,
		owner: owner,
	}
}

func (s *secretsManager) SetSecret(ctx context.Context, key, value string) error {
	ev, pkid, err := s.encryptValue(ctx, value)
	fmt.Println("Encrypted value:", ev)
	if err != nil {
		return err
	}

	return s.sendSecretRequest(ctx, pkid, key, ev)
}

func (s *secretsManager) encryptValue(ctx context.Context, value string) (string, string, error) {
	pk, pkid, err := s.getPublicKey(ctx)
	if err != nil {
		return "", "", err
	}

	data, err := exec.Command("node", "../../script.js", pk, value).Output()
	if err != nil {
		return "", "", err
	}

	fmt.Println("Encrypted value:", string(data))

	return string(data), pkid, nil
}

func (s *secretsManager) sendSecretRequest(ctx context.Context, pkid, name, value string) error {
	es := &github.EncryptedSecret{
		Name:           name,
		EncryptedValue: strings.Trim(value, "\n"),
		KeyID:          pkid,
	}

	res, err := s.c.Actions.CreateOrUpdateRepoSecret(ctx, s.owner, s.repo, es)
	if err != nil {
		return err
	}

	if res.StatusCode != 201 {
		if res.StatusCode != 204 {
			return errors.New("could not create or update secret")
		}
	}

	return nil
}