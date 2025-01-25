package lsp

import (
	"path/filepath"
	"strings"
)

// UriToPath converts a file:// URI to a path
func uriToPath(uri string) string {
	if strings.HasPrefix(uri, "file://") {
		return uri[7:]
	}
	return uri
}

// PathToUri converts a path to a file:// URI
func pathToUri(path string) string {
	if strings.HasPrefix(path, "file://") {
		return path
	}
	return "file://" + path
}

func getCanonicalPath(path string) string {
	path = filepath.Clean(path)

	resolvedPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		//fs.logger.Err(err)
	} else {
		path = resolvedPath
	}

	return path
}

func normalizePath(pathOrUri string) string {
	path := uriToPath(pathOrUri)
	return getCanonicalPath(path)
}

func replacePathFilename(path string, filename string) string {
	return filepath.Join(filepath.Dir(path), filename)
}