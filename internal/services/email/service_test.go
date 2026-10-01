package email

import (
	"bytes"
	"encoding/base64"
	"io"
	"mime"
	"mime/multipart"
	"net/mail"
	"strings"
	"testing"
)

func TestBuildAttachmentMessage(t *testing.T) {
	// Non-ASCII text, an equals sign and a line longer than the 76-character limit
	htmlBody := "<html><body>\n<p>Експорт даних: a=b</p>\n<p>" + strings.Repeat("long line ", 20) + "</p>\n</body></html>\n"
	attachment := bytes.Repeat([]byte{0x00, 0xff, 0x10, 0x7f}, 100)

	message, err := buildAttachmentMessage("from@example.com", "to@example.com", "Export", []byte(htmlBody), attachment, "export.xlsx")
	if err != nil {
		t.Fatalf("buildAttachmentMessage() error = %v", err)
	}

	parsed, err := mail.ReadMessage(bytes.NewReader(message))
	if err != nil {
		t.Fatalf("message is not parseable: %v", err)
	}
	mediaType, params, err := mime.ParseMediaType(parsed.Header.Get("Content-Type"))
	if err != nil || mediaType != "multipart/mixed" {
		t.Fatalf("Content-Type = %q (%v), want multipart/mixed", mediaType, err)
	}

	// The raw HTML part must really be quoted-printable: 7-bit only, lines within the limit
	rawParts := bytes.Split(message, []byte("--"+params["boundary"]))
	if len(rawParts) != 4 {
		t.Fatalf("message has %d boundary-separated chunks, want 4", len(rawParts))
	}
	for _, line := range strings.Split(string(rawParts[1]), "\r\n") {
		if len(line) > 76 {
			t.Errorf("HTML part line is %d characters long, limit is 76", len(line))
		}
		for _, c := range []byte(line) {
			if c > 127 {
				t.Fatalf("HTML part contains a raw 8-bit byte in line %q", line)
			}
		}
	}

	reader := multipart.NewReader(parsed.Body, params["boundary"])

	// multipart decodes quoted-printable parts transparently
	htmlPart, err := reader.NextPart()
	if err != nil {
		t.Fatalf("HTML part: %v", err)
	}
	decodedHTML, err := io.ReadAll(htmlPart)
	if err != nil {
		t.Fatalf("HTML part is not valid quoted-printable: %v", err)
	}
	wantHTML := strings.ReplaceAll(htmlBody, "\n", "\r\n")
	if got := strings.TrimSuffix(string(decodedHTML), "\r\n"); got != strings.TrimSuffix(wantHTML, "\r\n") {
		t.Errorf("decoded HTML body = %q, want %q", got, wantHTML)
	}

	attachmentPart, err := reader.NextPart()
	if err != nil {
		t.Fatalf("attachment part: %v", err)
	}
	if got := attachmentPart.FileName(); got != "export.xlsx" {
		t.Errorf("attachment file name = %q, want %q", got, "export.xlsx")
	}
	encodedAttachment, err := io.ReadAll(attachmentPart)
	if err != nil {
		t.Fatalf("attachment part: %v", err)
	}
	decodedAttachment, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(string(encodedAttachment), "\r\n", ""))
	if err != nil {
		t.Fatalf("attachment is not valid base64: %v", err)
	}
	if !bytes.Equal(decodedAttachment, attachment) {
		t.Error("decoded attachment differs from the original")
	}

	if _, err := reader.NextPart(); err != io.EOF {
		t.Errorf("expected end of message after the attachment, got %v", err)
	}
}
