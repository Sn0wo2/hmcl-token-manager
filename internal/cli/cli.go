package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"charm.land/huh/v2"
	"github.com/Sn0wo2/hmcl-token-manager/internal/account"
	"github.com/Sn0wo2/hmcl-token-manager/internal/utils"
	"github.com/atotto/clipboard"
)

func Run(path string) error {
	store, err := account.Load(filepath.Join(path, "private", "user-account-private-data.json"))
	if err != nil {
		return err
	}

	var action string
	if err := huh.NewSelect[string]().
		Title("HMCL Token Manager").
		Description(path).
		Options(
			huh.NewOption("Export account", "export"),
			huh.NewOption("Import account", "import"),
			huh.NewOption("Exit", "exit"),
		).
		Value(&action).
		Run(); err != nil {
		return err
	}

	switch action {
	case "export":
		if len(store.Accounts) == 0 {
			return huh.NewNote().
				Title("No accounts").
				Description("No HMCL account private data was found.").
				Next(true).
				Run()
		}

		options := make([]huh.Option[int], len(store.Accounts))
		for i, acc := range store.Accounts {
			name := acc.PrivateData.ProfileName
			if name == "" {
				name = acc.PrivateData.UserID
			}
			options[i] = huh.NewOption(name, i)
		}

		var selected int
		if err := huh.NewSelect[int]().
			Title("Select account").
			Description("Choose the Minecraft username to export").
			Options(options...).
			Value(&selected).
			Run(); err != nil {
			return err
		}

		var password string
		if err := huh.NewInput().
			Title("Encryption password").
			Description("This password is required when importing the token").
			EchoMode(huh.EchoModePassword).
			Validate(func(value string) error {
				if value == "" {
					return errors.New("password cannot be empty")
				}
				return nil
			}).
			Value(&password).
			Run(); err != nil {
			return err
		}

		encoded, err := utils.Crypto{}.Encrypt(store.Accounts[selected], password)
		if err != nil {
			return err
		}
		if err := clipboard.WriteAll(encoded); err != nil {
			_, err = fmt.Fprintln(os.Stdout, encoded)
			return err
		}
		return huh.NewNote().
			Title("Export complete").
			Description("Encrypted account data copied to clipboard.").
			Next(true).
			NextLabel("Done").
			Run()
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
				if err := huh.NewConfirm().
					Title("Use clipboard data?").
					Description(preview).
					Affirmative("Use").
					Negative("Paste manually").
					Value(&useClipboard).
					Run(); err != nil {
					return err
				}
	
				if useClipboard {
					encoded = value
				}
			}
		}
	
		if encoded == "" {
			if err := huh.NewText().
				Title("Import account").
				Description("Paste the encrypted HMCL account data").
				Lines(6).
				Validate(func(value string) error {
					if strings.TrimSpace(value) == "" {
						return errors.New("data cannot be empty")
					}
					return nil
				}).
				Value(&encoded).
				Run(); err != nil {
				return err
			}
		}
	
		encoded = strings.TrimSpace(encoded)
	
		var password string
		if err := huh.NewInput().
			Title("Decryption password").
			EchoMode(huh.EchoModePassword).
			Validate(func(value string) error {
				if value == "" {
					return errors.New("password cannot be empty")
				}
				return nil
			}).
			Value(&password).
			Run(); err != nil {
			return err
		}
	
		acc, err := utils.Crypto{}.Decrypt[account.Account](encoded, password)
		if err != nil {
			return err
		}
	
		added, err := store.Import(acc)
		if err != nil {
			return err
		}
	
		status := "Updated"
		if added {
			status = "Added"
		}
	
		return huh.NewNote().
			Title("Import complete").
			Description(fmt.Sprintf("%s %s\n\n%s", status, acc.PrivateData.ProfileName, path)).
			Next(true).
			NextLabel("Done").
			Run()
	}
	return nil
}
