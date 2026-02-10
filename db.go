package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

var db *pgx.Conn

func init() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found, relying on environment variables")
	}

}

func connectDB() error {
	DATABASE_URL, exists := os.LookupEnv("DATABASE_URL")
	if !exists {
		return fmt.Errorf("DATABASE_URL environment variable not set")
	}
	conn, err := pgx.Connect(context.Background(), DATABASE_URL)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	db = conn
	return nil
}

func lookupDestination(host, slug string) (string, error) {
	var dest string
	err := db.QueryRow(
		context.Background(),
		`
		SELECT l.destination_url
		FROM links l
		JOIN domains d ON d.hostname = l.domain_hostname
		WHERE d.hostname = $1
		  AND l.slug = $2
		  AND d.status = 'active'
		`,
		host,
		slug,
	).Scan(&dest)

	return dest, err
}

func upsertLink(domain, slug, destination string) error {
	_, err := db.Exec(
		context.Background(),
		`
		INSERT INTO links (domain_hostname, slug, destination_url)
		VALUES ($1, $2, $3)
		ON CONFLICT (domain_hostname, slug)
		DO UPDATE SET destination_url = EXCLUDED.destination_url
		`,
		domain,
		slug,
		destination,
	)
	return err
}

func deleteLink(domain, slug string) error {
	_, err := db.Exec(
		context.Background(),
		`
		DELETE FROM links
		WHERE domain_hostname = $1 AND slug = $2
		`,
		domain,
		slug,
	)
	return err
}

func upsertDomain(hostname, status string) error {
	_, err := db.Exec(
		context.Background(),
		`
		INSERT INTO domains (hostname, status)
		VALUES ($1, $2)
		ON CONFLICT (hostname)
		DO UPDATE SET status = EXCLUDED.status
		`,
		hostname,
		status,
	)
	return err
}

func deactivateDomain(hostname string) error {
	_, err := db.Exec(
		context.Background(),
		`
		UPDATE domains
		SET status = 'inactive'
		WHERE hostname = $1
		`,
		hostname,
	)
	return err
}
