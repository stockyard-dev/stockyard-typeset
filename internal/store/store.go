package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type DB struct { db *sql.DB }

type Document struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	Content      string   `json:"content"`
	Slug         string   `json:"slug"`
	Published    string   `json:"published"`
	CreatedAt    string   `json:"created_at"`
}

func Open(dataDir string) (*DB, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}
	dsn := filepath.Join(dataDir, "typeset.db") + "?_journal_mode=WAL&_busy_timeout=5000"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS documents (
			id TEXT PRIMARY KEY,\n\t\t\ttitle TEXT DEFAULT '',\n\t\t\tcontent TEXT DEFAULT '',\n\t\t\tslug TEXT DEFAULT '',\n\t\t\tpublished TEXT DEFAULT 'false',
			created_at TEXT DEFAULT (datetime('now'))
		)`)
	if err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return &DB{db: db}, nil
}

func (d *DB) Close() error { return d.db.Close() }

func genID() string { return fmt.Sprintf("%d", time.Now().UnixNano()) }

func (d *DB) Create(e *Document) error {
	e.ID = genID()
	e.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	_, err := d.db.Exec(`INSERT INTO documents (id, title, content, slug, published, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
		e.ID, e.Title, e.Content, e.Slug, e.Published, e.CreatedAt)
	return err
}

func (d *DB) Get(id string) *Document {
	row := d.db.QueryRow(`SELECT id, title, content, slug, published, created_at FROM documents WHERE id=?`, id)
	var e Document
	if err := row.Scan(&e.ID, &e.Title, &e.Content, &e.Slug, &e.Published, &e.CreatedAt); err != nil {
		return nil
	}
	return &e
}

func (d *DB) List() []Document {
	rows, err := d.db.Query(`SELECT id, title, content, slug, published, created_at FROM documents ORDER BY created_at DESC`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var result []Document
	for rows.Next() {
		var e Document
		if err := rows.Scan(&e.ID, &e.Title, &e.Content, &e.Slug, &e.Published, &e.CreatedAt); err != nil {
			continue
		}
		result = append(result, e)
	}
	return result
}

func (d *DB) Delete(id string) error {
	_, err := d.db.Exec(`DELETE FROM documents WHERE id=?`, id)
	return err
}

func (d *DB) Count() int {
	var n int
	d.db.QueryRow(`SELECT COUNT(*) FROM documents`).Scan(&n)
	return n
}
