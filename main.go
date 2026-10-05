package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Sn0wo2/hmcl-obfuscated"
	"github.com/google/uuid"
)

type UserAccountPrivateData struct {
	// account:${UUIDv7}
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

func main() {
	decrypt := false

	args := os.Args[1:]
	if len(args) > 0 {
		switch args[0] {
		case "decrypt":
			decrypt = true
		default:
			fmt.Println("Usage: hmcl-token-manager [decrypt]")
		}
	}

	path := filepath.Join(
		os.Getenv("APPDATA"),
		".hmcl",
		"private",
		"user-account-private-data.json",
	)

	fmt.Println("Reading ", path)

	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	var env hmcl.EnvelopeV1
	if err := json.Unmarshal(data, &env); err != nil {
		panic(err)
	}

	if env.Protection != "plain" {
		fmt.Println("user-account-private-data.json is encrypted, trying to decrypt it...")
		if err := env.Decrypt(); err != nil {
			panic(err)
		}
		fmt.Println("user-account-private-data.json decrypted successfully")
	}

	var accs []UserAccountPrivateData
	if err := json.Unmarshal(env.Payload, &accs); err != nil {
		panic(err)
	}

	if decrypt {
		fmt.Println("Decrypting user-account-private-data.json:")
		fmt.Println(env.Payload)
		os.Exit(0)
	}
	

	fmt.Print("Paste your hmcl account: ")
	var account UserAccountPrivateData

	if err := json.NewDecoder(os.Stdin).Decode(&account); err != nil {
		panic(err)
	}

	found := false

	for _, acc := range accs {
		if acc.PrivateData.UserID == account.PrivateData.UserID {
			fmt.Println("user-account-private-data.json already contains the account with the same userid, updating it...")
			acc.PrivateData = account.PrivateData
			found = true
			break
		}
	}

	if !found {
		fmt.Println("user-account-private-data.json does not contain the account with the same userid, adding it...")
		accID, err := uuid.NewV7()
		if err != nil {
			panic(err)
		}
		account.AccountID = "account:" + accID.String()
		accs = append(accs, account)
	}
	fmt.Println("user-account-private-data.json updated successfully, encrypting it...")

	if err := env.Encrypt(); err != nil {
		panic(err)
	}

	fmt.Println("Writing to ", path)

	envPayload, err := json.Marshal(env)
	if err != nil {
		panic(err)
	}

	os.WriteFile(path, envPayload, 0644)
}
