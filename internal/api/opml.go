package api

import (
	"encoding/xml"
	"io"
	"net/http"
	"time"
)

type OPML struct {
	XMLName xml.Name `xml:"opml"`
	Version string   `xml:"version,attr"`
	Head    OPMLHead `xml:"head"`
	Body    OPMLBody `xml:"body"`
}

type OPMLHead struct {
	Title       string `xml:"title"`
	DateCreated string `xml:"dateCreated,omitempty"`
}

type OPMLBody struct {
	Outlines []OPMLOutline `xml:"outline"`
}

type OPMLOutline struct {
	Text     string        `xml:"text,attr"`
	Title    string        `xml:"title,attr,omitempty"`
	Type     string        `xml:"type,attr,omitempty"`
	XMLURL   string        `xml:"xmlUrl,attr,omitempty"`
	HTMLURL  string        `xml:"htmlUrl,attr,omitempty"`
	Outlines []OPMLOutline `xml:"outline,omitempty"`
}

func (s *Server) handleImportOPML(w http.ResponseWriter, r *http.Request) {
	claims := getUserFromContext(r)

	file, _, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to read file")
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to read file content")
		return
	}

	var opml OPML
	if err := xml.Unmarshal(data, &opml); err != nil {
		writeError(w, http.StatusBadRequest, "invalid OPML format")
		return
	}

	imported := 0
	var importOutlines func(outlines []OPMLOutline, folderID *int64)
	importOutlines = func(outlines []OPMLOutline, folderID *int64) {
		for _, outline := range outlines {
			if outline.XMLURL != "" {
				// This is a feed
				_, err := s.db.CreateFeed(claims.UserID, outline.XMLURL, outline.Title, outline.HTMLURL, "", folderID)
				if err == nil {
					imported++
				}
			} else if len(outline.Outlines) > 0 {
				// This is a folder
				folder, err := s.db.CreateFolder(claims.UserID, outline.Text, nil)
				if err == nil {
					importOutlines(outline.Outlines, &folder.ID)
				}
			}
		}
	}

	importOutlines(opml.Body.Outlines, nil)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":  "import completed",
		"imported": imported,
	})
}

func (s *Server) handleExportOPML(w http.ResponseWriter, r *http.Request) {
	claims := getUserFromContext(r)

	feeds, err := s.db.GetFeeds(claims.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get feeds")
		return
	}

	folders, err := s.db.GetFolders(claims.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get folders")
		return
	}

	// Build folder map
	folderMap := make(map[int64]*OPMLOutline)
	var rootOutlines []OPMLOutline

	for _, folder := range folders {
		outline := &OPMLOutline{
			Text:     folder.Name,
			Title:    folder.Name,
			Outlines: []OPMLOutline{},
		}
		folderMap[folder.ID] = outline
		if folder.ParentID == nil {
			rootOutlines = append(rootOutlines, *outline)
		}
	}

	// Add feeds to folders or root
	for _, feed := range feeds {
		feedOutline := OPMLOutline{
			Text:    feed.Title,
			Title:   feed.Title,
			Type:    "rss",
			XMLURL:  feed.URL,
			HTMLURL: feed.SiteURL,
		}

		if feed.FolderID != nil {
			if folder, ok := folderMap[*feed.FolderID]; ok {
				folder.Outlines = append(folder.Outlines, feedOutline)
			}
		} else {
			rootOutlines = append(rootOutlines, feedOutline)
		}
	}

	opml := OPML{
		Version: "2.0",
		Head: OPMLHead{
			Title:       "RSS Reader Export",
			DateCreated: time.Now().Format(time.RFC1123),
		},
		Body: OPMLBody{
			Outlines: rootOutlines,
		},
	}

	w.Header().Set("Content-Type", "application/xml")
	w.Header().Set("Content-Disposition", "attachment; filename=subscriptions.opml")
	xml.NewEncoder(w).Encode(opml)
}
