package mhtmlparser

import (
	"bufio"
	"errors"
	//"bytes"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/google/uuid"
)

// Resource represents an extracted resource from an MHTML file.
type Resource struct {
	Type     string
	Filename string
	Data     []byte
	Size     int
	Source   string // embedded, inline, external
}

// MHTMLParser handles parsing and extraction of MHTML file resources.
type MHTMLParser struct {
	InputFile     string
	FetchExternal bool
	HTMLContent   string
	Resources     []Resource
	client        *http.Client // For external resource fetching
}

// New creates a new MHTMLParser instance.
func New(inputFile string, fetchExternal bool) *MHTMLParser {
	return &MHTMLParser{
		InputFile:     inputFile,
		FetchExternal: fetchExternal,
		client: &http.Client{
			Timeout: 5 * time.Second, // Set timeout for HTTP requests
		},
	}
}

// Parse reads and parses the MHTML file, extracting embedded resources and HTML content.
func (p *MHTMLParser) Parse() error {
	file, err := os.Open(p.InputFile)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	header, err := textproto.NewReader(reader).ReadMIMEHeader()
	if err != nil {
		return fmt.Errorf("failed to read MIME header: %w", err)
	}

	contentType := header.Get("Content-Type")
	_, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return fmt.Errorf("failed to parse media type: %w", err)
	}

	mr := multipart.NewReader(reader, params["boundary"])
	p.Resources = []Resource{}

	for {
		part, err := mr.NextPart()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			fmt.Printf("Warning: failed to read MIME part: %v\n", err)
			continue
		}

		contentType := normalizeContentType(part.Header.Get("Content-Type"))
		filename := part.FileName()
		if filename == "" {
			filename = fmt.Sprintf("resource_%s%s", randomID(), extensionFor(contentType))
		} else {
			filename = sanitizeFilename(filename)
		}

		data, err := io.ReadAll(part)
		if err != nil {
			fmt.Printf("Warning: failed to read part %s: %v\n", filename, err)
			continue
		}

		if strings.HasPrefix(contentType, "text/html") {
			p.HTMLContent = string(data)
			filename = fmt.Sprintf("page_%s.html", randomID())
			p.Resources = append(p.Resources, Resource{
				Type:     contentType,
				Filename: filename,
				Data:     data,
				Size:     len(data),
				Source:   "embedded",
			})
		} else {
			p.Resources = append(p.Resources, Resource{
				Type:     contentType,
				Filename: filename,
				Data:     data,
				Size:     len(data),
				Source:   "embedded",
			})
		}
	}

	// Extract inline and external JavaScript
	if p.HTMLContent != "" {
		if scripts, err := p.extractInlineScripts(); err == nil {
			p.Resources = append(p.Resources, scripts...)
		} else {
			fmt.Printf("Warning: failed to extract inline scripts: %v\n", err)
		}
		if p.FetchExternal {
			if scripts, err := p.downloadExternalScripts(); err == nil {
				p.Resources = append(p.Resources, scripts...)
			} else {
				fmt.Printf("Warning: failed to download external scripts: %v\n", err)
			}
		}
	}

	return nil
}

// extractInlineScripts extracts inline JavaScript from <script> tags without src attributes.
func (p *MHTMLParser) extractInlineScripts() ([]Resource, error) {
	var results []Resource

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(p.HTMLContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	doc.Find("script").Each(func(i int, s *goquery.Selection) {
		if _, exists := s.Attr("src"); !exists {
			code := strings.TrimSpace(s.Text())
			if code != "" {
				data := []byte(code)
				results = append(results, Resource{
					Type:     "text/javascript",
					Filename: fmt.Sprintf("inline_script_%s.js.random", randomID()),
					Data:     data,
					Size:     len(data),
					Source:   "inline",
				})
			}
		}
	})
	return results, nil
}

// downloadExternalScripts downloads external JavaScript from <script src="..."> tags.
func (p *MHTMLParser) downloadExternalScripts() ([]Resource, error) {
	var results []Resource
	scriptRe := regexp.MustCompile(`(?i)<script[^>]+src=["'](https?://[^"']+)["']`)

	matches := scriptRe.FindAllStringSubmatch(p.HTMLContent, -1)
	for _, match := range matches {
		url := match[1]
		resp, err := p.client.Get(url)
		if err != nil {
			fmt.Printf("Warning: failed to download %s: %v\n", url, err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			fmt.Printf("Warning: failed to download %s: status %d\n", url, resp.StatusCode)
			continue
		}

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Warning: failed to read response body for %s: %v\n", url, err)
			continue
		}

		name := sanitizeFilename(filepath.Base(strings.Split(url, "?")[0]))
		if name == "" || !strings.HasSuffix(name, ".js") {
			name = fmt.Sprintf("script_%s.js", randomID())
		}

		results = append(results, Resource{
			Type:     "text/javascript",
			Filename: name,
			Data:     data,
			Size:     len(data),
			Source:   "external",
		})
	}
	return results, nil
}

// ExtractResources saves resources to the output directory.
func (p *MHTMLParser) ExtractResources(outputDir string, selected []int) ([]string, error) {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Validate selected indices
	selectedSet := make(map[int]struct{})
	for _, idx := range selected {
		if idx < 0 || idx >= len(p.Resources) {
			return nil, fmt.Errorf("invalid selected index: %d", idx)
		}
		selectedSet[idx] = struct{}{}
	}

	var paths []string
	for i, res := range p.Resources {
		if selected != nil && !contains(selectedSet, i) {
			continue
		}
		outputPath := filepath.Join(outputDir, res.Filename)
		// Avoid collisions
		base := strings.TrimSuffix(outputPath, filepath.Ext(outputPath))
		ext := filepath.Ext(outputPath)
		counter := 1
		for fileExists(outputPath) {
			outputPath = fmt.Sprintf("%s_%d%s", base, counter, ext)
			counter++
		}
		if err := os.WriteFile(outputPath, res.Data, 0644); err != nil {
			return nil, fmt.Errorf("failed to write %s: %w", outputPath, err)
		}
		paths = append(paths, outputPath)
	}
	return paths, nil
}

// GetHTMLContent returns the HTML content of the MHTML file.
func (p *MHTMLParser) GetHTMLContent() string {
	return p.HTMLContent
}

// normalizeContentType normalizes content types by converting to lowercase and stripping parameters.
func normalizeContentType(contentType string) string {
	if contentType == "" {
		return "application/octet-stream"
	}
	return strings.ToLower(strings.Split(contentType, ";")[0])
}

// sanitizeFilename ensures filenames are safe for the filesystem.
func sanitizeFilename(name string) string {
	name = strings.ReplaceAll(name, string(os.PathSeparator), "_")
	name = regexp.MustCompile(`[^a-zA-Z0-9._-]`).ReplaceAllString(name, "_")
	if name == "" || name == "." || name == ".." {
		return fmt.Sprintf("resource_%s", randomID())
	}
	return name
}

// randomID generates a unique 8-character ID.
func randomID() string {
	return uuid.New().String()[:8]
}

// extensionFor maps content types to file extensions.
func extensionFor(contentType string) string {
	switch contentType {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	case "image/gif":
		return ".gif"
	case "text/css":
		return ".css"
	case "text/javascript", "application/javascript":
		return ".js"
	case "application/json":
		return ".json"
	case "font/ttf":
		return ".ttf"
	case "font/otf":
		return ".otf"
	case "font/woff":
		return ".woff"
	case "font/woff2":
		return ".woff2"
	case "text/plain":
		return ".txt"
	case "text/html":
		return ".html"
	default:
		return ".bin"
	}
}

// fileExists checks if a file exists.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// contains checks if a value exists in a set.
func contains(set map[int]struct{}, val int) bool {
	_, ok := set[val]
	return ok
}
