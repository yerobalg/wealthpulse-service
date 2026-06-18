// Command genpassword prints a bcrypt hash for a plaintext password using the
// same algorithm and cost the application uses (helper/cryptolib.Password).
//
// Use it to generate the value for the seeded admin user (or any user) in the
// migration under docs/sql.
//
// Usage:
//
//	go run ./scripts/genpassword -password=admin123 -cost=8
//
// -cost defaults to the PASSWORD_SALT_ROUND env var, or 10 if unset.
package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	password := flag.String("password", "", "plaintext password to hash (required)")
	cost := flag.Int("cost", defaultCost(), "bcrypt cost / salt round (matches PASSWORD_SALT_ROUND)")
	flag.Parse()

	if *password == "" {
		fmt.Fprintln(os.Stderr, "error: -password is required")
		flag.Usage()
		os.Exit(1)
	}

	if *cost < bcrypt.MinCost || *cost > bcrypt.MaxCost {
		fmt.Fprintf(os.Stderr, "error: -cost must be between %d and %d\n", bcrypt.MinCost, bcrypt.MaxCost)
		os.Exit(1)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(*password), *cost)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error: failed to hash password:", err)
		os.Exit(1)
	}

	fmt.Println(string(hash))
}

// defaultCost reads PASSWORD_SALT_ROUND, falling back to 10 when unset/invalid.
func defaultCost() int {
	if v, err := strconv.Atoi(os.Getenv("PASSWORD_SALT_ROUND")); err == nil {
		return v
	}
	return 10
}
