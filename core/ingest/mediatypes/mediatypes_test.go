package mediatypes

import "testing"

func TestResolve(t *testing.T) {
	tests := []struct {
		name     string
		sniffed  string
		filename string
		wantMIME string
		wantOK   bool
		wantReas Reason
	}{
		{name: "jpeg", sniffed: "image/jpeg", filename: "a.bin", wantMIME: "image/jpeg", wantOK: true},
		{name: "pdf", sniffed: "application/pdf", filename: "deed.pdf", wantMIME: "application/pdf", wantOK: true},
		{name: "plain", sniffed: "text/plain; charset=utf-8", filename: "notes.txt", wantMIME: "text/plain", wantOK: true},
		{name: "csv via plain sniff", sniffed: "text/plain", filename: "people.csv", wantMIME: "text/csv", wantOK: true},
		{name: "markdown via plain sniff", sniffed: "text/plain", filename: "notes.md", wantMIME: "text/markdown", wantOK: true},
		{name: "csv sniff", sniffed: "text/csv", filename: "people.csv", wantMIME: "text/csv", wantOK: true},
		{name: "mp3 via ext", sniffed: "application/octet-stream", filename: "oral.mp3", wantMIME: "audio/mpeg", wantOK: true},
		{name: "m4a via ext", sniffed: "application/octet-stream", filename: "clip.m4a", wantMIME: "audio/mp4", wantOK: true},
		{name: "unidentified", sniffed: "application/octet-stream", filename: "blob.xyz", wantReas: ReasonUnidentified},
		{name: "html sniff", sniffed: "text/html", filename: "page.html", wantReas: ReasonDisallowedSniff},
		{name: "docx zip", sniffed: "application/zip", filename: "notes.docx", wantMIME: "application/vnd.openxmlformats-officedocument.wordprocessingml.document", wantOK: true},
		{name: "doc mime", sniffed: "application/msword", filename: "letter.doc", wantMIME: "application/msword", wantOK: true},
		{name: "xlsx zip as office", sniffed: "application/zip", filename: "table.xlsx", wantReas: ReasonOffice},
		{name: "zip archive", sniffed: "application/zip", filename: "scans.zip", wantReas: ReasonArchive},
		{name: "exe", sniffed: "application/x-msdownload", filename: "setup.exe", wantReas: ReasonExecutable},
		{name: "excel mime", sniffed: "application/vnd.ms-excel", filename: "a.xls", wantReas: ReasonOffice},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, reason, ok := Resolve(tt.sniffed, tt.filename)
			if ok != tt.wantOK {
				t.Fatalf("ok=%v want %v (mime=%q reason=%q)", ok, tt.wantOK, got, reason)
			}
			if tt.wantOK {
				if got != tt.wantMIME {
					t.Fatalf("mime %q want %q", got, tt.wantMIME)
				}
				if reason != ReasonOK {
					t.Fatalf("reason %q", reason)
				}
				return
			}
			if reason != tt.wantReas {
				t.Fatalf("reason %q want %q", reason, tt.wantReas)
			}
		})
	}
}

func TestAllowed(t *testing.T) {
	if !Allowed("IMAGE/PNG") {
		t.Fatal("png")
	}
	if Allowed("application/zip") {
		t.Fatal("zip")
	}
}

func TestClassifyFilename(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want Reason
	}{
		{name: "pdf", in: "a.PDF", want: ReasonOK},
		{name: "docx", in: "a.docx", want: ReasonOK},
		{name: "doc", in: "a.doc", want: ReasonOK},
		{name: "xlsx", in: "a.xlsx", want: ReasonOffice},
		{name: "zip", in: "a.zip", want: ReasonArchive},
		{name: "exe", in: "a.exe", want: ReasonExecutable},
		{name: "unknown", in: "a.xyz", want: ReasonGenericType},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ClassifyFilename(tt.in); got != tt.want {
				t.Fatalf("got %q want %q", got, tt.want)
			}
		})
	}
}

func TestShortLabel(t *testing.T) {
	if ShortLabel("text/html") != "HTML" {
		t.Fatal(ShortLabel("text/html"))
	}
}
