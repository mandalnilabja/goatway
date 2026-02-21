package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/mandalnilabja/goatway/internal/storage"
)

func ensureAdminPassword(store storage.Storage) error {
	hasPassword, err := store.HasAdminPassword()
	if err != nil {
		return fmt.Errorf("failed to check admin password: %w", err)
	}

	if hasPassword {
		return nil
	}

	fmt.Println()
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║              FIRST-TIME SETUP REQUIRED                     ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("No admin password configured. Please set one now.")
	fmt.Println("This password protects the Web UI and Admin API.")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("Enter admin password (alphanumeric, min 8 chars): ")
		password, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read password: %w", err)
		}
		password = strings.TrimSpace(password)

		if !isValidAdminPassword(password) {
			fmt.Println("❌ Password must be alphanumeric with at least 8 characters.")
			fmt.Println()
			continue
		}

		fmt.Print("Confirm password: ")
		confirm, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read confirmation: %w", err)
		}
		confirm = strings.TrimSpace(confirm)

		if password != confirm {
			fmt.Println("❌ Passwords do not match. Please try again.")
			fmt.Println()
			continue
		}

		hash, err := storage.HashPassword(password, storage.DefaultArgon2Params())
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}

		if err := store.SetAdminPasswordHash(hash); err != nil {
			return fmt.Errorf("failed to save password: %w", err)
		}

		fmt.Println()
		fmt.Println("✓ Admin password saved successfully!")
		fmt.Println()
		return nil
	}
}

func isValidAdminPassword(password string) bool {
	if len(password) < 8 {
		return false
	}
	for _, c := range password {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')) {
			return false
		}
	}
	return true
}
