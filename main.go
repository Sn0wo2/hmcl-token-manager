package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"charm.land/huh/v2"
	"github.com/Sn0wo2/caelum"
	"github.com/Sn0wo2/hmcl-token-manager/internal/account"
	"github.com/Sn0wo2/hmcl-token-manager/internal/utils"
	"github.com/atotto/clipboard"
)

//nolint:maintidx
func main() {
	check := func(err error) {
		if err == nil {
			return
		}

		caelum.New(caelum.Config{
			Targets: []caelum.Target{{Writer: os.Stderr}},
		}).Error("HMCL Token Manager failed", "error", err)
		os.Exit(1)
	}

	path := os.Getenv("HMCL_USER_HOME")
	if strings.TrimSpace(path) == "" {
		path = ""

		switch runtime.GOOS {
		case "windows":
			if appdata := os.Getenv("APPDATA"); strings.TrimSpace(appdata) != "" {
				path = filepath.Join(appdata, ".hmcl")
			}
		case "linux", "android", "freebsd", "solaris", "illumos":
			if data := os.Getenv("XDG_DATA_HOME"); strings.TrimSpace(data) != "" {
				path = filepath.Join(data, "hmcl")
			}
		}

		if path == "" {
			home, err := os.UserHomeDir()
			check(err)

			switch runtime.GOOS {
			case "windows":
				path = filepath.Join(home, ".hmcl")
			case "darwin", "ios":
				path = filepath.Join(home, "Library", "Application Support", "hmcl")
			case "linux", "android", "freebsd", "solaris", "illumos":
				path = filepath.Join(home, ".local", "share", "hmcl")
			default:
				path = filepath.Join(home, "hmcl")
			}
		}
	}

	path, err := filepath.Abs(path)
	check(err)

	if !filepath.IsAbs(path) {
		check(fmt.Errorf("HMCL path is not absolute: %s", path))
	}

	path = filepath.Clean(path)

	info, err := os.Stat(path) //nolint:gosec // HMCL_USER_HOME and platform-specific directories are intentional inputs.
	check(err)

	if !info.IsDir() {
		check(fmt.Errorf("HMCL path is not a directory: %s", path))
	}

	store, err := account.Load(path)
	check(err)

	var action string
	check(huh.NewSelect[string]().
		Title("HMCL Token Manager").
		Description(path).
		Options(
			huh.NewOption("Export account", "export"),
			huh.NewOption("Import account", "import"),
			huh.NewOption("Exit", "exit"),
		).
		Value(&action).
		Run())

	switch action {
	case "export":
		if len(store.Accounts) == 0 {
			check(huh.NewNote().
				Title("No accounts").
				Description("No HMCL account private data was found.").
				Next(true).
				Run())

			return
		}

		metadataByID := make(map[string]account.Metadata, len(store.UserAccounts.Accounts))
		for _, metadata := range store.UserAccounts.Accounts {
			if _, exists := metadataByID[metadata.AccountID]; !exists {
				metadataByID[metadata.AccountID] = metadata
			}
		}

		options := make([]huh.Option[int], len(store.Accounts))
		for i, acc := range store.Accounts {
			options[i] = huh.NewOption(utils.FirstNonEmpty(acc.PrivateData.ProfileName, metadataByID[acc.AccountID].ProfileName, &acc.AccountID), i)
		}

		var selected int
		check(huh.NewSelect[int]().
			Title("Select account").
			Description("Choose the Minecraft username to export").
			Options(options...).
			Value(&selected).
			Run())

		var password string
		check(huh.NewInput().
			Title("Encryption password").
			Description("This password is required when importing the token").
			EchoMode(huh.EchoModePassword).
			Validate(huh.ValidateNotEmpty()).
			Value(&password).
			Run())

		acc := store.Accounts[selected]

		metadata, ok := metadataByID[acc.AccountID]
		if !ok {
			check(fmt.Errorf("account metadata not found: %s", acc.AccountID))
		}

		encoded, err := utils.Encrypt(account.Transfer{Account: acc, Metadata: metadata}, password)
		check(err)

		if err := clipboard.WriteAll(encoded); err != nil {
			_, err = fmt.Fprintln(os.Stdout, encoded)
			check(err)

			return
		}

		check(huh.NewNote().
			Title("Export complete").
			Description("Encrypted account data copied to clipboard.").
			Next(true).
			NextLabel("Done").
			Run())
	case "import":
		var encoded string

		if value, err := clipboard.ReadAll(); err == nil {
			value = strings.TrimSpace(value)

			if value != "" {
				preview := value
				if len(preview) > 64 {
					preview = preview[:64] + "..."
				}

				var useClipboard bool
				check(huh.NewConfirm().
					Title("Use clipboard data?").
					Description(preview).
					Affirmative("Use").
					Negative("Paste manually").
					Value(&useClipboard).
					Run())

				if useClipboard {
					encoded = value
				}
			}
		}

		if encoded == "" {
			check(huh.NewText().
				Title("Import account").
				Description("Paste the encrypted HMCL account data").
				Lines(6).
				Validate(func(value string) error {
					return huh.ValidateNotEmpty()(strings.TrimSpace(value))
				}).
				Value(&encoded).
				Run())
		}

		var password string
		check(huh.NewInput().
			Title("Decryption password").
			EchoMode(huh.EchoModePassword).
			Validate(huh.ValidateNotEmpty()).
			Value(&password).
			Run())

		transfer, err := utils.Decrypt[account.Transfer](strings.TrimSpace(encoded), password)
		check(err)

		added, err := store.Import(transfer)
		check(err)

		status := "Updated"
		if added {
			status = "Added"
		}

		check(huh.NewNote().
			Title("Import complete").
			Description(fmt.Sprintf("%s %s\n\n%s", status, utils.FirstNonEmpty(transfer.Account.PrivateData.ProfileName, transfer.Metadata.ProfileName, &transfer.Account.AccountID), path)).
			Next(true).
			NextLabel("Done").
			Run())
	}
}
