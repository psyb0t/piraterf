package server

import (
	"encoding/json"
	"html/template"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/psyb0t/aichteeteapee"
)

type DirectoryIndexingType uint

const (
	DirectoryIndexingTypeNone DirectoryIndexingType = iota // No directory indexing (default: 0)
	DirectoryIndexingTypeHTML                              // HTML directory listing
	DirectoryIndexingTypeJSON                              // JSON directory listing
)

type DirectoryEntry struct {
	Name    string    `json:"name"`
	IsDir   bool      `json:"isDir"`
	Size    int64     `json:"size"`
	ModTime time.Time `json:"modTime"`
	URL     string    `json:"url"`
}

type DirectoryListing struct {
	Path    string           `json:"path"`
	Parent  string           `json:"parent,omitempty"`
	Entries []DirectoryEntry `json:"entries"`
}

// HTML template for directory listings
const directoryListingHTMLTemplate = `<!DOCTYPE html>
<html>
<head>
    <title>Index of {{.Path}}</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .container { max-width: 1000px; }
        .header { border-bottom: 1px solid #ccc; margin-bottom: 20px; padding-bottom: 10px; }
        .parent { margin-bottom: 20px; }
        .parent a { text-decoration: none; color: #0066cc; font-weight: bold; }
        .entries { border-collapse: collapse; width: 100%; }
        .entries th { text-align: left; padding: 8px; border-bottom: 2px solid #ddd; background: #f5f5f5; }
        .entries td { padding: 8px; border-bottom: 1px solid #eee; }
        .entries tr:hover { background-color: #f9f9f9; }
        .name { width: 50%; }
        .name a { text-decoration: none; color: #0066cc; }
        .name a:hover { text-decoration: underline; }
        .dir { font-weight: bold; }
        .dir:before { content: "📁 "; }
        .file:before { content: "📄 "; }
        .size { text-align: right; width: 100px; }
        .date { width: 200px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Index of {{.Path}}</h1>
        </div>

        {{if .Parent}}
        <div class="parent">
            <a href="{{.Parent}}">[Parent Directory]</a>
        </div>
        {{end}}

        <table class="entries">
            <thead>
                <tr>
                    <th class="name">Name</th>
                    <th class="size">Size</th>
                    <th class="date">Last Modified</th>
                </tr>
            </thead>
            <tbody>
                {{range .Entries}}
                <tr>
                    <td class="name">
                        <a href="{{.URL}}" class="{{if .IsDir}}dir{{else}}file{{end}}">{{.Name}}</a>
                    </td>
                    <td class="size">
                        {{if .IsDir}}-{{else}}{{.Size}} bytes{{end}}
                    </td>
                    <td class="date">{{.ModTime.Format "2006-01-02 15:04:05"}}</td>
                </tr>
                {{end}}
            </tbody>
        </table>
    </div>
</body>
</html>`

// handleDirectoryRequest handles directory access - tries to serve index.html or provides directory listing
func (s *Server) handleDirectoryRequest(
	w http.ResponseWriter,
	r *http.Request,
	fullPath string,
	staticConfig StaticRouteConfig,
) {
	// Try to serve index.html from the directory first
	indexPath := filepath.Join(fullPath, aichteeteapee.FileNameIndexHTML)
	if _, err := os.Stat(indexPath); err == nil {
		// index.html exists, serve it directly
		http.ServeFile(w, r, indexPath)

		return
	}

	// No index.html found - check if directory indexing is enabled
	if staticConfig.DirectoryIndexingType == DirectoryIndexingTypeNone {
		// Directory listing disabled, return forbidden
		aichteeteapee.WriteJSON(
			w,
			http.StatusForbidden,
			aichteeteapee.ErrorResponseDirectoryListingNotSupported,
		)

		return
	}

	// Directory indexing is enabled - generate listing
	s.generateDirectoryListing(w, r, fullPath, staticConfig)
}

// generateDirectoryListing creates and serves a directory listing
func (s *Server) generateDirectoryListing(
	w http.ResponseWriter,
	r *http.Request,
	fullPath string,
	staticConfig StaticRouteConfig,
) {
	// Read directory contents
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		s.logger.WithError(err).Error("failed to read directory for listing")
		aichteeteapee.WriteJSON(
			w,
			http.StatusInternalServerError,
			aichteeteapee.ErrorResponseInternalServerError,
		)

		return
	}

	// Build directory listing
	listing := s.buildDirectoryListing(r, entries, staticConfig)

	// Serve the listing in the requested format
	switch staticConfig.DirectoryIndexingType {
	case DirectoryIndexingTypeNone:
		// This should never happen since we check for DirectoryIndexingNone earlier,
		// but adding for exhaustive lint compliance
		aichteeteapee.WriteJSON(
			w,
			http.StatusForbidden,
			aichteeteapee.ErrorResponseDirectoryListingNotSupported,
		)

		return
	case DirectoryIndexingTypeJSON:
		w.Header().Set(aichteeteapee.HeaderNameContentType, aichteeteapee.ContentTypeJSON)

		// For JSON, just return the entries array directly
		if err := json.NewEncoder(w).Encode(listing.Entries); err != nil {
			s.logger.WithError(err).Error("failed to encode directory listing JSON")
			aichteeteapee.WriteJSON(
				w,
				http.StatusInternalServerError,
				aichteeteapee.ErrorResponseInternalServerError,
			)
		}
	case DirectoryIndexingTypeHTML:
		fallthrough
	default:
		w.Header().Set(aichteeteapee.HeaderNameContentType, "text/html; charset=utf-8")

		tmpl, err := template.New("directory").Parse(directoryListingHTMLTemplate)
		if err != nil {
			s.logger.WithError(err).Error("failed to parse directory listing template")
			aichteeteapee.WriteJSON(
				w,
				http.StatusInternalServerError,
				aichteeteapee.ErrorResponseInternalServerError,
			)

			return
		}

		if err := tmpl.Execute(w, listing); err != nil {
			s.logger.WithError(err).Error("failed to execute directory listing template")
		}
	}
}

// buildDirectoryListing creates a DirectoryListing from directory contents
func (s *Server) buildDirectoryListing(
	r *http.Request,
	entries []os.DirEntry,
	staticConfig StaticRouteConfig,
) DirectoryListing {
	relativePath := strings.TrimPrefix(r.URL.Path, staticConfig.Path)
	if relativePath == "" {
		relativePath = "/"
	}

	listing := DirectoryListing{
		Path:    relativePath,
		Entries: make([]DirectoryEntry, 0, len(entries)),
	}

	if relativePath != "/" {
		listing.Parent = s.buildParentURL(relativePath, staticConfig)
	}

	// Process directory entries
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue // Skip entries we can't stat
		}

		entryPath := filepath.Join(relativePath, entry.Name())

		// Build URL properly using path.Join for URL construction
		var entryURL string
		if staticConfig.Path == "/" {
			entryURL = entryPath
		} else {
			// Use path.Join to properly construct the URL path
			entryURL = path.Join(staticConfig.Path, entryPath)
		}

		if entry.IsDir() && !strings.HasSuffix(entryURL, "/") {
			entryURL += "/"
		}

		dirEntry := DirectoryEntry{
			Name:    entry.Name(),
			IsDir:   entry.IsDir(),
			Size:    info.Size(),
			ModTime: info.ModTime(),
			URL:     entryURL,
		}

		listing.Entries = append(listing.Entries, dirEntry)
	}

	// Sort entries: directories first, then files, both alphabetically
	sort.Slice(listing.Entries, func(i, j int) bool {
		if listing.Entries[i].IsDir != listing.Entries[j].IsDir {
			return listing.Entries[i].IsDir // Directories first
		}

		return strings.ToLower(listing.Entries[i].Name) < strings.ToLower(listing.Entries[j].Name)
	})

	return listing
}

func (s *Server) buildParentURL(
	relativePath string,
	staticConfig StaticRouteConfig,
) string {
	cleanRelativePath := strings.TrimSuffix(relativePath, "/")
	parentPath := path.Dir(cleanRelativePath)

	if parentPath == "." {
		parentPath = "/"
	}

	var parentURL string

	switch {
	case staticConfig.Path == "/":
		parentURL = parentPath
	case parentPath == "/":
		parentURL = staticConfig.Path
	default:
		parentURL = path.Join(staticConfig.Path, parentPath)
	}

	if !strings.HasSuffix(parentURL, "/") {
		parentURL += "/"
	}

	return parentURL
}
