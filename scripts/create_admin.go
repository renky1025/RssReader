package main

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	password := "admin"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Password hash for 'admin': %s\n", string(hash))
	fmt.Printf("\nRun this SQL:\n")
	fmt.Printf("INSERT INTO users (username, password_hash, is_admin) VALUES ('admin', '%s', 1);\n", string(hash))
}
