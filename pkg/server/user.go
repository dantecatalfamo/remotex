package server

import (
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
)

const BearerTokenByteLength = 32

// CreateUser adds a user to the database and creates their directory
func CreateUser(config Config, name string) error {
	if _, err := config.database.conn.Exec("INSERT INTO users (name) VALUES (?)", name); err != nil {
		return fmt.Errorf("CreateUser insert in db: %w", err)
	}

	userDir := filepath.Join(config.ProjectDir, name)
	if err := os.Mkdir(userDir, 0700); err != nil {
		return fmt.Errorf("CreateUser make user dir: %w", err)
	}

	return nil
}

// DeleteUser deletes a user from the database and recursively removes their directory
func DeleteUser(config Config, name string) error {
	if _, err := config.database.conn.Exec("DELETE FROM users WHERE name = ?", name); err != nil {
		return fmt.Errorf("DeleteUser delete from db: %w", err)
	}

	userDir := filepath.Join(config.ProjectDir, name)
	if err := os.RemoveAll(userDir); err != nil {
		return fmt.Errorf("DeleteUser RemoveAll dir: %w", err)
	}

	return nil
}

// CreateUserToken generates a new random token for a user and stores
// it in the database. It retuens the newly generated token
func CreateUserToken(config Config, userName, tokenDescription string) (string, error) {
	userId, err := config.database.GetUserId(userName)
	if err != nil {
		return "", fmt.Errorf("CreateUserToken get user id: %w", err)
	}

	buffer := make([]byte, BearerTokenByteLength)
	if _, err := rand.Read(buffer); err != nil {
		return "", fmt.Errorf("CreateUserToken read random: %w", err)
	}

	token := fmt.Sprintf("%x", buffer)

	if _, err := config.database.conn.Exec(
		"INSERT INTO tokens (user_id, token, description) VALUES (?, ?, ?)",
		userId,
		token,
		tokenDescription,
	); err != nil {
		return "", fmt.Errorf("CreateUserToken insert db: %w", err)
	}

	return token, nil
}

func DeleteUserToken(config Config, token string) error {
	if _, err := config.database.conn.Exec("DELETE FROM tokens WHERE token = ?", token); err != nil {
		return fmt.Errorf("DeleteUserToken exec: %w", err)
	}
	return nil
}

func GetUserFromToken(config Config, token string) (string, error) {
	row := config.database.conn.QueryRow("SELECT u.name FROM users u JOIN tokens t ON u.id = t.user_id WHERE t.token = ? LIMIT 1", token)
	if row.Err() != nil {
		return "", fmt.Errorf("GetUserIdFromToken query: %w", row.Err())
	}
	var user string
	if err := row.Scan(&user); err != nil {
		return "", fmt.Errorf("GetUserFromToken scan: %w", err)
	}

	return user, nil
}
