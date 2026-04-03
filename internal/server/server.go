package server

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/stockyard-dev/stockyard-typeset/internal/store"
)

type Server struct { db *store.DB; mux *http.ServeMux }

func New(db *store.DB, limits Limits) *Server {
	s := &Server{db: db, mux: http.NewServeMux(), limits: limits}
	s.mux.HandleFunc("GET /api/sites", s.listSites)
	s.mux.HandleFunc("POST /api/sites", s.createSite)
	s.mux.HandleFunc("GET /api/sites/{id}", s.getSite)
	s.mux.HandleFunc("PUT /api/sites/{id}", s.updateSite)
	s.mux.HandleFunc("DELETE /api/sites/{id}", s.deleteSite)
	s.mux.HandleFunc("GET /api/sites/{id}/nav", s.navigation)
	s.mux.HandleFunc("GET /api/sites/{id}/search", s.search)

	s.mux.HandleFunc("POST /api/sections", s.createSection)
	s.mux.HandleFunc("GET /api/sites/{id}/sections", s.listSections)
	s.mux.HandleFunc("DELETE /api/sections/{id}", s.deleteSection)

	s.mux.HandleFunc("GET /api/pages", s.listPages)
	s.mux.HandleFunc("POST /api/pages", s.createPage)
	s.mux.HandleFunc("GET /api/pages/{id}", s.getPage)
	s.mux.HandleFunc("GET /api/pages/{id}/toc", s.pageTOC)
	s.mux.HandleFunc("PUT /api/pages/{id}", s.updatePage)
	s.mux.HandleFunc("DELETE /api/pages/{id}", s.deletePage)

	s.mux.HandleFunc("GET /api/stats", s.stats)
	s.mux.HandleFunc("GET /api/health", s.health)

	s.mux.HandleFunc("GET /ui", s.dashboard)
	s.mux.HandleFunc("GET /ui/", s.dashboard)
	s.mux.HandleFunc("GET /", s.root)
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Public docs: /docs/{site-slug}/{page-slug}
	if strings.HasPrefix(r.URL.Path, "/docs/") {
		s.renderDoc(w, r); return
	}
	s.mux.ServeHTTP(w, r)
}

func writeJSON(w http.ResponseWriter, code int, v any) { w.Header().Set("Content-Type","application/json"); w.WriteHeader(code); json.NewEncoder(w).Encode(v) }
func writeErr(w http.ResponseWriter, code int, msg string) { writeJSON(w, code, map[string]string{"error": msg}) }
func (s *Server) root(w http.ResponseWriter, r *http.Request) { if r.URL.Path != "/" { http.NotFound(w, r); return }; http.Redirect(w, r, "/ui", http.StatusFound) }

func (s *Server) listSites(w http.ResponseWriter, r *http.Request) { writeJSON(w, 200, map[string]any{"sites": orEmpty(s.db.ListSites())}) }
func (s *Server) createSite(w http.ResponseWriter, r *http.Request) {
	var site store.Site; json.NewDecoder(r.Body).Decode(&site)
	if site.Name == "" { writeErr(w, 400, "name required"); return }
	s.db.CreateSite(&site); writeJSON(w, 201, s.db.GetSite(site.ID))
}
func (s *Server) getSite(w http.ResponseWriter, r *http.Request) { site := s.db.GetSite(r.PathValue("id")); if site == nil { writeErr(w, 404, "not found"); return }; writeJSON(w, 200, site) }
func (s *Server) updateSite(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id"); ex := s.db.GetSite(id); if ex == nil { writeErr(w, 404, "not found"); return }
	var site store.Site; json.NewDecoder(r.Body).Decode(&site)
	if site.Name == "" { site.Name = ex.Name }; if site.Slug == "" { site.Slug = ex.Slug }
	if site.Version == "" { site.Version = ex.Version }
	s.db.UpdateSite(id, &site); writeJSON(w, 200, s.db.GetSite(id))
}
func (s *Server) deleteSite(w http.ResponseWriter, r *http.Request) { s.db.DeleteSite(r.PathValue("id")); writeJSON(w, 200, map[string]string{"deleted":"ok"}) }
func (s *Server) navigation(w http.ResponseWriter, r *http.Request) { writeJSON(w, 200, map[string]any{"nav": orEmpty(s.db.Navigation(r.PathValue("id")))}) }
func (s *Server) search(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]any{"pages": orEmpty(s.db.SearchPages(r.PathValue("id"), r.URL.Query().Get("q")))})
}

func (s *Server) createSection(w http.ResponseWriter, r *http.Request) {
	var sec store.Section; json.NewDecoder(r.Body).Decode(&sec)
	if sec.Name == "" || sec.SiteID == "" { writeErr(w, 400, "name and site_id required"); return }
	s.db.CreateSection(&sec); writeJSON(w, 201, sec)
}
func (s *Server) listSections(w http.ResponseWriter, r *http.Request) { writeJSON(w, 200, map[string]any{"sections": orEmpty(s.db.ListSections(r.PathValue("id")))}) }
func (s *Server) deleteSection(w http.ResponseWriter, r *http.Request) { s.db.DeleteSection(r.PathValue("id")); writeJSON(w, 200, map[string]string{"deleted":"ok"}) }

func (s *Server) listPages(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]any{"pages": orEmpty(s.db.ListPages(r.URL.Query().Get("site_id"), r.URL.Query().Get("section_id")))})
}
func (s *Server) createPage(w http.ResponseWriter, r *http.Request) {
	var p store.Page; json.NewDecoder(r.Body).Decode(&p)
	if p.Title == "" || p.SiteID == "" { writeErr(w, 400, "title and site_id required"); return }
	s.db.CreatePage(&p); writeJSON(w, 201, s.db.GetPage(p.ID))
}
func (s *Server) getPage(w http.ResponseWriter, r *http.Request) { p := s.db.GetPage(r.PathValue("id")); if p == nil { writeErr(w, 404, "not found"); return }; writeJSON(w, 200, p) }
func (s *Server) pageTOC(w http.ResponseWriter, r *http.Request) {
	p := s.db.GetPage(r.PathValue("id")); if p == nil { writeErr(w, 404, "not found"); return }
	writeJSON(w, 200, map[string]any{"toc": store.GenerateTOC(p.Body)})
}
func (s *Server) updatePage(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id"); ex := s.db.GetPage(id); if ex == nil { writeErr(w, 404, "not found"); return }
	var p store.Page; json.NewDecoder(r.Body).Decode(&p)
	if p.Title == "" { p.Title = ex.Title }; if p.Slug == "" { p.Slug = ex.Slug }
	if p.SiteID == "" { p.SiteID = ex.SiteID }
	s.db.UpdatePage(id, &p); writeJSON(w, 200, s.db.GetPage(id))
}
func (s *Server) deletePage(w http.ResponseWriter, r *http.Request) { s.db.DeletePage(r.PathValue("id")); writeJSON(w, 200, map[string]string{"deleted":"ok"}) }

// ── Public doc rendering ──

func (s *Server) renderDoc(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/docs/"), "/")
	if len(parts) < 1 { http.NotFound(w, r); return }
	siteSlug := parts[0]
	pageSlug := ""
	if len(parts) >= 2 { pageSlug = parts[1] }

	site := s.db.GetSiteBySlug(siteSlug)
	if site == nil { http.NotFound(w, r); return }

	if pageSlug == "" {
		// Serve navigation/index
		nav := s.db.Navigation(site.ID)
		writeJSON(w, 200, map[string]any{"site": site, "nav": orEmpty(nav)})
		return
	}

	page := s.db.GetPageBySlug(siteSlug, pageSlug)
	if page == nil || page.Draft { http.NotFound(w, r); return }
	toc := store.GenerateTOC(page.Body)
	writeJSON(w, 200, map[string]any{"site": site, "page": page, "toc": toc})
}

func (s *Server) stats(w http.ResponseWriter, r *http.Request) { writeJSON(w, 200, s.db.Stats()) }
func (s *Server) health(w http.ResponseWriter, r *http.Request) { st := s.db.Stats(); writeJSON(w, 200, map[string]any{"status":"ok","service":"typeset","sites":st.Sites,"pages":st.Pages}) }
func orEmpty[T any](s []T) []T { if s == nil { return []T{} }; return s }
func init() { log.SetFlags(log.LstdFlags | log.Lshortfile) }
