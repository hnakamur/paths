// path utility package
package paths

import (
	"io/ioutil"
	"os"
	"path"
	"time"
)

// Read directory entries recursively with depth-first order.
// Entries in each directory are sorted by names.
// Entries in a directory follows the directory.
// Name() for an entry returns a path starting with dir.
// If matcher is specified, only entries which matches will be returned.
// If marker is specified, entries after maker will be returned.
// If maxEntries is greater than zero, the number of entries will be limited.
func RecurReadDir(dir string, matcher Matcher, marker string, maxEntries int) (entries []os.FileInfo, err error) {
	r := &recurDirReader{dir, matcher, marker, maxEntries, ioutil.ReadDir}
	return r.recurReadDir()
}

type recurDirReader struct {
	dir         string
	matcher     Matcher
	marker      string
	maxEntries  int
	readDirFunc func(string) ([]os.FileInfo, error)
}

func (r *recurDirReader) recurReadDir() ([]os.FileInfo, error) {
	entries := make([]os.FileInfo, 0)
	if r.marker != "" {
		return r.appendReadDirAfterMarker(r.marker, entries)
	} else {
		return r.appendReadDir(r.dir, entries)
	}
}

func (r *recurDirReader) appendReadDir(dirname string, entries []os.FileInfo) ([]os.FileInfo, error) {
	infos, err := r.readDirFunc(dirname)
	if err != nil {
		return entries, err
	}

	return r.processFileInfos(dirname, infos, entries)
}

func (r *recurDirReader) appendReadDirAfterMarker(marker string, entries []os.FileInfo) ([]os.FileInfo, error) {
	markerBase := path.Base(marker)
	markerDir := path.Dir(marker)
	infos, err := r.readDirFunc(markerDir)
	if err != nil {
		return entries, err
	}

	i := indexOfName(infos, markerBase)

	if marker == r.marker && infos[i].IsDir() {
		subinfos, err := r.readDirFunc(marker)
		if err != nil {
			return entries, err
		}
		entries, err = r.processFileInfos(marker, subinfos, entries)
		if err != nil || r.hasReachedLimit(entries) {
			return entries, err
		}
	}

	entries, err = r.processFileInfos(markerDir, infos[i+1:], entries)
	if err != nil || r.hasReachedLimit(entries) {
		return entries, err
	}

	for ; markerDir != r.dir; markerDir = path.Dir(markerDir) {
		entries, err = r.appendReadDirAfterMarker(markerDir, entries)
		if err != nil || r.hasReachedLimit(entries) {
			return entries, err
		}
	}
	return entries, nil
}

func indexOfName(infos []os.FileInfo, name string) int {
	i := 0
	for ; i < len(infos); i++ {
		if infos[i].Name() == name {
			break
		}
	}
	return i
}

func (r *recurDirReader) processFileInfos(dirname string, infos []os.FileInfo,
	entries []os.FileInfo) ([]os.FileInfo, error) {
	var err error
	for _, info := range infos {
		entryPath := path.Join(dirname, info.Name())
		if r.matcher == nil || r.matcher.Match(entryPath) {
			entry := &dirEntry{
				entryPath,
				info.Size(),
				info.Mode(),
				info.ModTime(),
				info.Sys()}
			entries = append(entries, entry)
			if r.hasReachedLimit(entries) {
				return entries, nil
			}
		}

		if info.IsDir() {
			subdir := path.Join(dirname, info.Name())
			entries, err = r.appendReadDir(subdir, entries)
			if err != nil || r.hasReachedLimit(entries) {
				return entries, err
			}
		}
	}
	return entries, nil
}

func (r *recurDirReader) hasReachedLimit(entries []os.FileInfo) bool {
	return r.maxEntries > 0 && len(entries) >= r.maxEntries
}

type dirEntry struct {
	name    string      // the file name with the relative direcotry
	size    int64       // length in bytes for regular files; system-dependent for others
	mode    os.FileMode // file mode bits
	modTime time.Time   // modification time
	sys     interface{} // underlying data source (can return nil)
}

func (e *dirEntry) Name() string       { return e.name }
func (e *dirEntry) Size() int64        { return e.size }
func (e *dirEntry) Mode() os.FileMode  { return e.mode }
func (e *dirEntry) ModTime() time.Time { return e.modTime }
func (e *dirEntry) IsDir() bool        { return e.Mode().IsDir() }
func (e *dirEntry) Sys() interface{}   { return e.sys }
