// Saves a blob as a file: a temporary object URL and an <a download> click (CSP-compatible, no inline scripts)
export function downloadBlob(blob: Blob, filename: string) {
  const url = window.URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = filename;
  document.body.appendChild(a);
  a.click();
  window.URL.revokeObjectURL(url);
  document.body.removeChild(a);
}

// Extracts the file name from a Content-Disposition header (filename*=UTF-8''... or filename="..."),
// without any path part; null when the header is missing or has no usable name
export function filenameFromContentDisposition(header: string | null): string | null {
  if (!header) {
    return null;
  }

  let name: string | null = null;
  const extended = /filename\*\s*=\s*UTF-8''([^;]+)/i.exec(header);
  if (extended) {
    try {
      name = decodeURIComponent(extended[1].trim());
    } catch {
      // Malformed encoding: fall back to the plain filename parameter
    }
  }
  if (name === null) {
    const plain = /filename\s*=\s*(?:"([^"]*)"|([^;]+))/i.exec(header);
    name = plain ? (plain[1] ?? plain[2]).trim() : null;
  }

  const basename = name?.split(/[/\\]/).pop()?.trim();
  return basename ? basename : null;
}
