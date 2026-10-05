package account

import (
	"encoding/json"
	"fmt"
	"os"

	hmcl "github.com/Sn0wo2/hmcl-obfuscated"
	"github.com/google/uuid"
)

type Account struct {
	AccountID   string      `json:"accountID"`
	PrivateData PrivateData `json:"privateData"`
}

type PrivateData struct {
	ProfileName  string `json:"profileName"`
	TokenType    string `json:"tokenType"`
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	NotAfter     int64  `json:"notAfter"`
	UserID       string `json:"userid"`
}

type Store struct {
	Accounts []Account

	path     string
	envelope hmcl.EnvelopeV1
}

func Load(path string) (*Store, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read account file: %w", err)
	}

	var env hmcl.EnvelopeV1
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, fmt.Errorf("parse account file envelope: %w", err)
	}

	if env.Protection != "plain" {
		if err := env.Decrypt(); err != nil {
			return nil, fmt.Errorf("decrypt account file: %w", err)
		}
	}

	var accounts []Account
	if err := json.Unmarshal(env.Payload, &accounts); err != nil {
		return nil, fmt.Errorf("parse account file payload: %w", err)
	}

	return &Store{Accounts: accounts, path: path, envelope: env}, nil
}

func (s *Store) Import(account Account) (bool, error) {
	if account.PrivateData.UserID == "" {
		return false, fmt.Errorf("imported account has no userid")
	}

	for i := range s.Accounts {
		if s.Accounts[i].PrivateData.UserID == account.PrivateData.UserID {
			account.AccountID = s.Accounts[i].AccountID
			s.Accounts[i] = account

			return false, s.save()
		}
	}

	accountID, err := uuid.NewV7()
	if err != nil {
		return false, fmt.Errorf("generate account id: %w", err)
	}

	account.AccountID = "account:" + accountID.String()
	s.Accounts = append(s.Accounts, account)

	if err := s.save(); err != nil {
		return false, err
	}

	return true, nil
}

func (s *Store) save() error {
	payload, err := json.Marshal(s.Accounts)
	if err != nil {
		return fmt.Errorf("marshal accounts: %w", err)
	}

	s.envelope.Payload = payload

	if err := s.envelope.Encrypt(); err != nil {
		return fmt.Errorf("encrypt account file: %w", err)
	}

	output, err := json.Marshal(s.envelope)
	if err != nil {
		return fmt.Errorf("marshal account file: %w", err)
	}

	if err := os.WriteFile(s.path, output, 0600); err != nil {
		return fmt.Errorf("write account file: %w", err)
	}

	return nil
}
