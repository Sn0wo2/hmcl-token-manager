package account

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"os"
	"path/filepath"
	"slices"

	"github.com/Sn0wo2/hmcl-obfuscated"
)

type UserAccounts struct {
	Schema   string         `json:"$schema"`
	Accounts []Metadata     `json:"accounts,omitzero"`
	Extra    jsontext.Value `json:",embed"`
}

type Metadata struct {
	Type        string         `json:"type"`
	AccountID   string         `json:"accountID"`
	ProfileID   *string        `json:"profileID,omitzero"`
	ProfileName *string        `json:"profileName,omitzero"`
	Extra       jsontext.Value `json:",embed"`
}

type Transfer struct {
	Account  hmcl.Account `json:"account"`
	Metadata Metadata     `json:"metadata"`
}

type Store struct {
	Accounts     []hmcl.Account
	UserAccounts UserAccounts

	path     string
	envelope hmcl.EnvelopeV1
}

func Load(path string) (*Store, error) {
	data, err := os.ReadFile(filepath.Join(path, "private", "user-account-private-data.json"))
	if err != nil {
		return nil, err
	}

	var envelope hmcl.EnvelopeV1
	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil, err
	}

	accounts, err := envelope.Decrypt(func(data []byte, v any) error { return json.Unmarshal(data, v) })
	if err != nil {
		return nil, err
	}

	data, err = os.ReadFile(filepath.Join(path, "config", "user-accounts.json"))
	if err != nil {
		return nil, err
	}
	var userAccounts UserAccounts
	if err := json.Unmarshal(data, &userAccounts); err != nil {
		return nil, err
	}

	return &Store{Accounts: accounts, UserAccounts: userAccounts, path: path, envelope: envelope}, nil
}

func (s *Store) Import(data Transfer) (bool, error) {
	if data.Account.AccountID == "" || data.Account.AccountID != data.Metadata.AccountID || data.Metadata.Type == "" {
		return false, errors.New("invalid account transfer: matching account IDs and account type are required")
	}

	i := slices.IndexFunc(s.Accounts, func(acc hmcl.Account) bool { return acc.AccountID == data.Account.AccountID })
	if i >= 0 {
		s.Accounts[i] = data.Account
	} else {
		s.Accounts = append(s.Accounts, data.Account)
	}
	i = slices.IndexFunc(s.UserAccounts.Accounts, func(acc Metadata) bool { return acc.AccountID == data.Account.AccountID })
	if i >= 0 {
		s.UserAccounts.Accounts[i] = data.Metadata
	} else {
		s.UserAccounts.Accounts = append(s.UserAccounts.Accounts, data.Metadata)
	}

	return i < 0, s.save()
}

func (s *Store) save() error {
	metadata, err := json.Marshal(s.UserAccounts)
	if err != nil {
		return err
	}
	envelope, err := hmcl.Encrypt(s.Accounts, func(v any) ([]byte, error) { return json.Marshal(v) })
	if err != nil {
		return err
	}
	envelope.Schema, envelope.Extra = s.envelope.Schema, s.envelope.Extra
	output, err := json.Marshal(envelope)
	if err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(s.path, "private", "user-account-private-data.json"), output, 0600); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(s.path, "config", "user-accounts.json"), metadata, 0600)
}
