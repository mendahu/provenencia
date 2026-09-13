// Package mediatypes is the ingest allowlist and reject-reason classifier.
// Go remains authoritative; clients may mirror for fail-fast UX.
package mediatypes

import (
	"path/filepath"
	"strings"
)

// Reason is a stable reject taxonomy shared with macOS Callout mapping.
type Reason string

const (
	ReasonOK              Reason = ""
	ReasonOffice          Reason = "office"
	ReasonArchive         Reason = "archive"
	ReasonExecutable      Reason = "executable"
	ReasonEmpty           Reason = "empty"
	ReasonTooLarge        Reason = "too_large"
	ReasonUnidentified    Reason = "unidentified"
	ReasonDisallowedSniff Reason = "disallowed_sniff"
	ReasonGenericType     Reason = "generic_type"
)

// Normalize strips parameters and lowercases a MIME type.
func Normalize(mediaType string) string {
	mediaType = strings.TrimSpace(mediaType)
	if i := strings.IndexByte(mediaType, ';'); i >= 0 {
		mediaType = mediaType[:i]
	}
	return strings.ToLower(strings.TrimSpace(mediaType))
}

// Allowed reports whether mediaType is accepted for new ingest.
func Allowed(mediaType string) bool {
	_, ok := allowed[Normalize(mediaType)]
	return ok
}

// Resolve picks the stored media type from a sniff result and original filename.
// When sniff is application/octet-stream, a controlled extension map may upgrade
// to an allowed AV/text type. When sniff is text/plain, a more specific allowed
// text/* from the extension (csv, markdown) is preferred so Files are not
// stored as generic TXT.
func Resolve(sniffed, originalFilename string) (mediaType string, reason Reason, ok bool) {
	sniffed = Normalize(sniffed)
	if sniffed == "" {
		sniffed = "application/octet-stream"
	}

	if Allowed(sniffed) {
		if sniffed == "text/plain" {
			if mapped, mappedOK := fromExtension(originalFilename); mappedOK && mapped != "text/plain" {
				return mapped, ReasonOK, true
			}
		}
		return sniffed, ReasonOK, true
	}

	if sniffed == "application/octet-stream" {
		if mapped, mappedOK := fromExtension(originalFilename); mappedOK {
			return mapped, ReasonOK, true
		}
		return sniffed, ReasonUnidentified, false
	}

	// OOXML and similar are often sniffed as zip — prefer Office when the
	// extension says so, so Callouts can say “Word documents…” not “Archives…”.
	if sniffed == "application/zip" || sniffed == "application/x-zip-compressed" {
		if r := classifyExtension(originalFilename); r == ReasonOffice {
			return sniffed, ReasonOffice, false
		}
		return sniffed, ReasonArchive, false
	}

	if r := classifyMIME(sniffed); r != ReasonGenericType {
		return sniffed, r, false
	}
	if r := classifyExtension(originalFilename); r != ReasonGenericType {
		return sniffed, r, false
	}
	return sniffed, ReasonDisallowedSniff, false
}

// ClassifyFilename returns a reject reason from extension alone (Swift-style
// precheck). Empty when the extension is allowed or unknown-but-not-classified.
func ClassifyFilename(originalFilename string) Reason {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(originalFilename), "."))
	if ext == "" {
		return ReasonGenericType
	}
	if _, ok := extAllowed[ext]; ok {
		return ReasonOK
	}
	return classifyExtension(originalFilename)
}

var allowed = map[string]struct{}{
	"image/jpeg": {},
	"image/png":  {},
	"image/gif":  {},
	"image/webp": {},
	"image/tiff": {},
	"image/tif":  {},
	"image/bmp":  {},
	"image/heic": {},
	"image/heif": {},

	"application/pdf": {},
	"text/plain":      {},
	"text/csv":        {},
	"application/csv": {},
	"text/markdown":   {},
	"text/x-markdown": {},

	"audio/mpeg":  {},
	"audio/mp4":   {},
	"audio/x-m4a": {},
	"audio/aac":   {},
	"audio/wav":   {},
	"audio/wave":  {},
	"audio/x-wav": {},
	"audio/ogg":   {},
	"audio/flac":  {},
	"audio/aiff":  {},
	"audio/x-aiff": {},

	"video/mp4":       {},
	"video/quicktime": {},
	"video/webm":      {},
	"video/x-m4v":     {},
	"video/x-msvideo": {},
}

// extAllowed maps extension (no dot) → MIME for octet-stream fallback and
// client allow checks.
var extAllowed = map[string]string{
	"jpg":  "image/jpeg",
	"jpeg": "image/jpeg",
	"png":  "image/png",
	"gif":  "image/gif",
	"webp": "image/webp",
	"tif":  "image/tiff",
	"tiff": "image/tiff",
	"bmp":  "image/bmp",
	"heic": "image/heic",
	"heif": "image/heif",
	"pdf":  "application/pdf",
	"txt":  "text/plain",
	"text": "text/plain",
	"csv":  "text/csv",
	"md":   "text/markdown",
	"markdown": "text/markdown",
	"mp3":  "audio/mpeg",
	"m4a":  "audio/mp4",
	"aac":  "audio/aac",
	"wav":  "audio/wav",
	"ogg":  "audio/ogg",
	"flac": "audio/flac",
	"aif":  "audio/aiff",
	"aiff": "audio/aiff",
	"mp4":  "video/mp4",
	"m4v":  "video/x-m4v",
	"mov":  "video/quicktime",
	"webm": "video/webm",
	"avi":  "video/x-msvideo",
}

func fromExtension(originalFilename string) (string, bool) {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(originalFilename), "."))
	if ext == "" {
		return "", false
	}
	mime, ok := extAllowed[ext]
	return mime, ok
}

func classifyMIME(mime string) Reason {
	switch mime {
	case "application/msword",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"application/vnd.ms-excel",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		"application/vnd.ms-powerpoint",
		"application/vnd.openxmlformats-officedocument.presentationml.presentation",
		"application/rtf",
		"text/rtf",
		"application/vnd.oasis.opendocument.text",
		"application/vnd.oasis.opendocument.spreadsheet",
		"application/vnd.oasis.opendocument.presentation":
		return ReasonOffice
	case "application/zip",
		"application/x-zip-compressed",
		"application/x-rar-compressed",
		"application/vnd.rar",
		"application/x-7z-compressed",
		"application/x-tar",
		"application/gzip",
		"application/x-gzip",
		"application/x-bzip2",
		"application/x-apple-diskimage":
		return ReasonArchive
	case "application/x-msdownload",
		"application/x-dosexec",
		"application/x-executable",
		"application/x-mach-binary",
		"application/vnd.microsoft.portable-executable":
		return ReasonExecutable
	}
	if strings.HasPrefix(mime, "application/vnd.ms-") ||
		strings.HasPrefix(mime, "application/vnd.openxmlformats-officedocument.") {
		return ReasonOffice
	}
	return ReasonGenericType
}

func classifyExtension(originalFilename string) Reason {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(originalFilename), "."))
	switch ext {
	case "doc", "docx", "xls", "xlsx", "ppt", "pptx", "rtf", "odt", "ods", "odp":
		return ReasonOffice
	case "zip", "rar", "7z", "tar", "gz", "tgz", "bz2", "dmg":
		return ReasonArchive
	case "exe", "msi", "bat", "cmd", "com", "scr", "app", "sh", "bin", "dll", "so", "dylib":
		return ReasonExecutable
	default:
		return ReasonGenericType
	}
}

// ShortLabel returns a short human label for disallowed_sniff Callouts.
func ShortLabel(mediaType string) string {
	switch Normalize(mediaType) {
	case "text/html":
		return "HTML"
	case "text/xml", "application/xml":
		return "XML"
	case "image/svg+xml":
		return "SVG"
	case "application/json", "text/json":
		return "JSON"
	case "text/css":
		return "CSS"
	case "application/javascript", "text/javascript":
		return "JavaScript"
	default:
		if mediaType == "" {
			return "That"
		}
		return mediaType
	}
}
