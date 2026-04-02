package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type DB struct{ db *sql.DB }

type Site struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description,omitempty"`
	Version     string `json:"version,omitempty"`
	CreatedAt   string `json:"created_at"`
	PageCount   int    `json:"page_count"`
	SectionCount int   `json:"section_count"`
}

type Section struct {
	ID        string `json:"id"`
	SiteID    string `json:"site_id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	Position  int    `json:"position"`
	CreatedAt string `json:"created_at"`
	PageCount int    `json:"page_count"`
}

type Page struct {
	ID        string `json:"id"`
	SiteID    string `json:"site_id"`
	SectionID string `json:"section_id,omitempty"`
	Title     string `json:"title"`
	Slug      string `json:"slug"`
	Body      string `json:"body"`
	Position  int    `json:"position"`
	Draft     bool   `json:"draft"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	WordCount int    `json:"word_count"`
}

type TOCEntry struct {
	Level int    `json:"level"`
	Text  string `json:"text"`
	Slug  string `json:"slug"`
}

type NavItem struct {
	Section Section `json:"section"`
	Pages   []Page  `json:"pages"`
}

func Open(dataDir string) (*DB, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil { return nil, err }
	dsn := filepath.Join(dataDir, "typeset.db") + "?_journal_mode=WAL&_busy_timeout=5000"
	db, err := sql.Open("sqlite", dsn)
	if err != nil { return nil, err }
	for _, q := range []string{
		`CREATE TABLE IF NOT EXISTS sites (id TEXT PRIMARY KEY, name TEXT NOT NULL, slug TEXT UNIQUE NOT NULL, description TEXT DEFAULT '', version TEXT DEFAULT 'latest', created_at TEXT DEFAULT (datetime('now')))`,
		`CREATE TABLE IF NOT EXISTS sections (id TEXT PRIMARY KEY, site_id TEXT NOT NULL REFERENCES sites(id), name TEXT NOT NULL, slug TEXT DEFAULT '', position INTEGER DEFAULT 0, created_at TEXT DEFAULT (datetime('now')))`,
		`CREATE TABLE IF NOT EXISTS pages (id TEXT PRIMARY KEY, site_id TEXT NOT NULL REFERENCES sites(id), section_id TEXT DEFAULT '', title TEXT NOT NULL, slug TEXT DEFAULT '', body TEXT DEFAULT '', position INTEGER DEFAULT 0, draft INTEGER DEFAULT 0, created_at TEXT DEFAULT (datetime('now')), updated_at TEXT DEFAULT (datetime('now')))`,
		`CREATE INDEX IF NOT EXISTS idx_sections_site ON sections(site_id)`,
		`CREATE INDEX IF NOT EXISTS idx_pages_site ON pages(site_id)`,
		`CREATE INDEX IF NOT EXISTS idx_pages_section ON pages(section_id)`,
	} {
		if _, err := db.Exec(q); err != nil { return nil, fmt.Errorf("migrate: %w", err) }
	}
	return &DB{db: db}, nil
}

func (d *DB) Close() error { return d.db.Close() }
func genID() string { return fmt.Sprintf("%d", time.Now().UnixNano()) }
func now() string { return time.Now().UTC().Format(time.RFC3339) }

// ── Sites ──

func (d *DB) CreateSite(s *Site) error {
	s.ID = genID(); s.CreatedAt = now()
	if s.Slug == "" { s.Slug = strings.ToLower(strings.ReplaceAll(strings.TrimSpace(s.Name), " ", "-")) }
	if s.Version == "" { s.Version = "latest" }
	_, err := d.db.Exec(`INSERT INTO sites (id,name,slug,description,version,created_at) VALUES (?,?,?,?,?,?)`,
		s.ID, s.Name, s.Slug, s.Description, s.Version, s.CreatedAt)
	return err
}

func (d *DB) hydrateSite(s *Site) {
	d.db.QueryRow(`SELECT COUNT(*) FROM pages WHERE site_id=?`, s.ID).Scan(&s.PageCount)
	d.db.QueryRow(`SELECT COUNT(*) FROM sections WHERE site_id=?`, s.ID).Scan(&s.SectionCount)
}

func (d *DB) GetSite(id string) *Site {
	var s Site
	if err := d.db.QueryRow(`SELECT id,name,slug,description,version,created_at FROM sites WHERE id=?`, id).Scan(&s.ID, &s.Name, &s.Slug, &s.Description, &s.Version, &s.CreatedAt); err != nil { return nil }
	d.hydrateSite(&s); return &s
}

func (d *DB) GetSiteBySlug(slug string) *Site {
	var s Site
	if err := d.db.QueryRow(`SELECT id,name,slug,description,version,created_at FROM sites WHERE slug=?`, slug).Scan(&s.ID, &s.Name, &s.Slug, &s.Description, &s.Version, &s.CreatedAt); err != nil { return nil }
	d.hydrateSite(&s); return &s
}

func (d *DB) ListSites() []Site {
	rows, _ := d.db.Query(`SELECT id,name,slug,description,version,created_at FROM sites ORDER BY name`)
	if rows == nil { return nil }; defer rows.Close()
	var out []Site
	for rows.Next() {
		var s Site; rows.Scan(&s.ID, &s.Name, &s.Slug, &s.Description, &s.Version, &s.CreatedAt)
		d.hydrateSite(&s); out = append(out, s)
	}
	return out
}

func (d *DB) UpdateSite(id string, s *Site) error {
	_, err := d.db.Exec(`UPDATE sites SET name=?,slug=?,description=?,version=? WHERE id=?`, s.Name, s.Slug, s.Description, s.Version, id); return err
}

func (d *DB) DeleteSite(id string) error {
	d.db.Exec(`DELETE FROM pages WHERE site_id=?`, id)
	d.db.Exec(`DELETE FROM sections WHERE site_id=?`, id)
	_, err := d.db.Exec(`DELETE FROM sites WHERE id=?`, id); return err
}

// ── Sections ──

func (d *DB) CreateSection(s *Section) error {
	s.ID = genID(); s.CreatedAt = now()
	if s.Slug == "" { s.Slug = strings.ToLower(strings.ReplaceAll(strings.TrimSpace(s.Name), " ", "-")) }
	_, err := d.db.Exec(`INSERT INTO sections (id,site_id,name,slug,position,created_at) VALUES (?,?,?,?,?,?)`,
		s.ID, s.SiteID, s.Name, s.Slug, s.Position, s.CreatedAt)
	return err
}

func (d *DB) ListSections(siteID string) []Section {
	rows, _ := d.db.Query(`SELECT id,site_id,name,slug,position,created_at FROM sections WHERE site_id=? ORDER BY position, name`, siteID)
	if rows == nil { return nil }; defer rows.Close()
	var out []Section
	for rows.Next() {
		var s Section; rows.Scan(&s.ID, &s.SiteID, &s.Name, &s.Slug, &s.Position, &s.CreatedAt)
		d.db.QueryRow(`SELECT COUNT(*) FROM pages WHERE section_id=?`, s.ID).Scan(&s.PageCount)
		out = append(out, s)
	}
	return out
}

func (d *DB) DeleteSection(id string) error {
	d.db.Exec(`UPDATE pages SET section_id='' WHERE section_id=?`, id)
	_, err := d.db.Exec(`DELETE FROM sections WHERE id=?`, id); return err
}

// ── Pages ──

func (d *DB) CreatePage(p *Page) error {
	p.ID = genID(); p.CreatedAt = now(); p.UpdatedAt = p.CreatedAt
	if p.Slug == "" { p.Slug = strings.ToLower(strings.ReplaceAll(strings.TrimSpace(p.Title), " ", "-")) }
	p.WordCount = len(strings.Fields(p.Body))
	draft := 0; if p.Draft { draft = 1 }
	_, err := d.db.Exec(`INSERT INTO pages (id,site_id,section_id,title,slug,body,position,draft,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?,?,?)`,
		p.ID, p.SiteID, p.SectionID, p.Title, p.Slug, p.Body, p.Position, draft, p.CreatedAt, p.UpdatedAt)
	return err
}

func (d *DB) GetPage(id string) *Page {
	var p Page; var draft int
	if err := d.db.QueryRow(`SELECT id,site_id,section_id,title,slug,body,position,draft,created_at,updated_at FROM pages WHERE id=?`, id).Scan(
		&p.ID, &p.SiteID, &p.SectionID, &p.Title, &p.Slug, &p.Body, &p.Position, &draft, &p.CreatedAt, &p.UpdatedAt); err != nil { return nil }
	p.Draft = draft == 1; p.WordCount = len(strings.Fields(p.Body)); return &p
}

func (d *DB) GetPageBySlug(siteSlug, pageSlug string) *Page {
	site := d.GetSiteBySlug(siteSlug)
	if site == nil { return nil }
	var p Page; var draft int
	if err := d.db.QueryRow(`SELECT id,site_id,section_id,title,slug,body,position,draft,created_at,updated_at FROM pages WHERE site_id=? AND slug=?`, site.ID, pageSlug).Scan(
		&p.ID, &p.SiteID, &p.SectionID, &p.Title, &p.Slug, &p.Body, &p.Position, &draft, &p.CreatedAt, &p.UpdatedAt); err != nil { return nil }
	p.Draft = draft == 1; p.WordCount = len(strings.Fields(p.Body)); return &p
}

func (d *DB) ListPages(siteID, sectionID string) []Page {
	q := `SELECT id,site_id,section_id,title,slug,body,position,draft,created_at,updated_at FROM pages WHERE site_id=?`
	args := []any{siteID}
	if sectionID != "" { q += ` AND section_id=?`; args = append(args, sectionID) }
	q += ` ORDER BY position, title`
	rows, _ := d.db.Query(q, args...)
	if rows == nil { return nil }; defer rows.Close()
	var out []Page
	for rows.Next() {
		var p Page; var draft int
		rows.Scan(&p.ID, &p.SiteID, &p.SectionID, &p.Title, &p.Slug, &p.Body, &p.Position, &draft, &p.CreatedAt, &p.UpdatedAt)
		p.Draft = draft == 1; p.WordCount = len(strings.Fields(p.Body))
		out = append(out, p)
	}
	return out
}

func (d *DB) UpdatePage(id string, p *Page) error {
	p.UpdatedAt = now(); p.WordCount = len(strings.Fields(p.Body))
	draft := 0; if p.Draft { draft = 1 }
	_, err := d.db.Exec(`UPDATE pages SET title=?,slug=?,body=?,section_id=?,position=?,draft=?,updated_at=? WHERE id=?`,
		p.Title, p.Slug, p.Body, p.SectionID, p.Position, draft, p.UpdatedAt, id)
	return err
}

func (d *DB) DeletePage(id string) error { _, err := d.db.Exec(`DELETE FROM pages WHERE id=?`, id); return err }

func (d *DB) SearchPages(siteID, query string) []Page {
	s := "%" + query + "%"
	rows, _ := d.db.Query(`SELECT id,site_id,section_id,title,slug,body,position,draft,created_at,updated_at FROM pages WHERE site_id=? AND (title LIKE ? OR body LIKE ?) ORDER BY title`, siteID, s, s)
	if rows == nil { return nil }; defer rows.Close()
	var out []Page
	for rows.Next() {
		var p Page; var draft int
		rows.Scan(&p.ID, &p.SiteID, &p.SectionID, &p.Title, &p.Slug, &p.Body, &p.Position, &draft, &p.CreatedAt, &p.UpdatedAt)
		p.Draft = draft == 1; p.WordCount = len(strings.Fields(p.Body))
		out = append(out, p)
	}
	return out
}

// ── Navigation ──

func (d *DB) Navigation(siteID string) []NavItem {
	sections := d.ListSections(siteID)
	var nav []NavItem
	// Unsectioned pages first
	unsectioned := d.ListPages(siteID, "")
	// Only include pages without section
	var topPages []Page
	for _, p := range unsectioned {
		if p.SectionID == "" { topPages = append(topPages, p) }
	}
	if len(topPages) > 0 {
		nav = append(nav, NavItem{Section: Section{Name: "Getting Started", Slug: ""}, Pages: topPages})
	}
	for _, sec := range sections {
		pages := d.ListPages(siteID, sec.ID)
		nav = append(nav, NavItem{Section: sec, Pages: pages})
	}
	return nav
}

// ── TOC from markdown ──

func GenerateTOC(body string) []TOCEntry {
	var toc []TOCEntry
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "### ") {
			text := strings.TrimPrefix(line, "### ")
			toc = append(toc, TOCEntry{Level: 3, Text: text, Slug: strings.ToLower(strings.ReplaceAll(text, " ", "-"))})
		} else if strings.HasPrefix(line, "## ") {
			text := strings.TrimPrefix(line, "## ")
			toc = append(toc, TOCEntry{Level: 2, Text: text, Slug: strings.ToLower(strings.ReplaceAll(text, " ", "-"))})
		} else if strings.HasPrefix(line, "# ") {
			text := strings.TrimPrefix(line, "# ")
			toc = append(toc, TOCEntry{Level: 1, Text: text, Slug: strings.ToLower(strings.ReplaceAll(text, " ", "-"))})
		}
	}
	return toc
}

// ── Stats ──

type Stats struct { Sites int `json:"sites"`; Pages int `json:"pages"`; Sections int `json:"sections"`; Words int `json:"words"` }
func (d *DB) Stats() Stats {
	var s Stats
	d.db.QueryRow(`SELECT COUNT(*) FROM sites`).Scan(&s.Sites)
	d.db.QueryRow(`SELECT COUNT(*) FROM pages`).Scan(&s.Pages)
	d.db.QueryRow(`SELECT COUNT(*) FROM sections`).Scan(&s.Sections)
	rows, _ := d.db.Query(`SELECT body FROM pages`)
	if rows != nil { defer rows.Close(); for rows.Next() { var b string; rows.Scan(&b); s.Words += len(strings.Fields(b)) } }
	return s
}
