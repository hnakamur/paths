package paths

import (
	"os"
	"sort"
	"testing"
	"time"
)

type fakeFileInfo struct {
	dir      bool
	basename string
	ents     []*fakeFileInfo
}

func (e *fakeFileInfo) Name() string       { return e.basename }
func (e *fakeFileInfo) Size() int64        { return 0 }
func (e *fakeFileInfo) ModTime() time.Time { return time.Time{} }
func (e *fakeFileInfo) Sys() interface{}   { return nil }
func (e *fakeFileInfo) IsDir() bool        { return e.dir }
func (f *fakeFileInfo) Mode() os.FileMode {
	if f.dir {
		return 0755 | os.ModeDir
	}
	return 0644
}

type fakeFS map[string]*fakeFileInfo

func newFakeFS() fakeFS {
	archive_tar_testdata := &fakeFileInfo{true, "testdata", []*fakeFileInfo{
		&fakeFileInfo{false, "gnu.tar", nil},
		&fakeFileInfo{false, "pax.tar", nil},
		&fakeFileInfo{false, "small.txt", nil},
	}}
	archive_tar := &fakeFileInfo{true, "tar", []*fakeFileInfo{
		&fakeFileInfo{false, "common.go", nil},
		&fakeFileInfo{false, "example_test.go", nil},
		&fakeFileInfo{false, "reader.go", nil},
		&fakeFileInfo{false, "reader_test.go", nil},
		archive_tar_testdata,
		&fakeFileInfo{false, "writer.go", nil},
		&fakeFileInfo{false, "writer_test.go", nil},
	}}
	archive_zip_testdata := &fakeFileInfo{true, "testdata", []*fakeFileInfo{
		&fakeFileInfo{false, "crc32-not-streamed.zip", nil},
		&fakeFileInfo{false, "dd.zip", nil},
		&fakeFileInfo{false, "go-no-datadesc-sig.zip", nil},
	}}
	archive_zip := &fakeFileInfo{true, "zip", []*fakeFileInfo{
		&fakeFileInfo{false, "example_test.go", nil},
		&fakeFileInfo{false, "reader.go", nil},
		&fakeFileInfo{false, "reader_test.go", nil},
		&fakeFileInfo{false, "struct.go", nil},
		archive_zip_testdata,
		&fakeFileInfo{false, "writer.go", nil},
		&fakeFileInfo{false, "writer_test.go", nil},
		&fakeFileInfo{false, "zip_test.go", nil},
	}}
	archive := &fakeFileInfo{true, "archive", []*fakeFileInfo{
		archive_tar,
		archive_zip,
	}}
	return fakeFS{
		"archive":              archive,
		"archive/tar":          archive_tar,
		"archive/tar/testdata": archive_tar_testdata,
		"archive/zip":          archive_zip,
		"archive/zip/testdata": archive_zip_testdata,
	}
}

type byName []os.FileInfo

func (x byName) Len() int           { return len(x) }
func (x byName) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }
func (x byName) Less(i, j int) bool { return x[i].Name() < x[j].Name() }

func fakeReaderDirFunc(fs fakeFS) func(dirname string) ([]os.FileInfo, error) {
	return func(dirname string) ([]os.FileInfo, error) {
		fi, ok := fs[dirname]
		if !ok {
			return nil, os.ErrInvalid
		}

		ents := make([]os.FileInfo, len(fi.ents))
		for i, ent := range fi.ents {
			ents[i] = ent
		}
		sort.Sort(byName(ents))
		return ents, nil
	}
}

type resultFileInfo struct {
	dir  bool
	name string
}

func (e *resultFileInfo) Name() string       { return e.name }
func (e *resultFileInfo) Size() int64        { return 0 }
func (e *resultFileInfo) ModTime() time.Time { return time.Time{} }
func (e *resultFileInfo) Sys() interface{}   { return nil }
func (e *resultFileInfo) IsDir() bool        { return e.dir }
func (f *resultFileInfo) Mode() os.FileMode {
	if f.dir {
		return 0755 | os.ModeDir
	}
	return 0644
}

func checkFileInfo(t *testing.T, fi os.FileInfo, expected os.FileInfo) {
	if fi.IsDir() != expected.IsDir() {
		t.Errorf("%s: IsDir=%v, expected=%v", fi.Name(), fi.IsDir(), expected.IsDir())
	}
	if fi.Name() != expected.Name() {
		t.Errorf("%s: expected=%s", fi.Name(), expected.Name())
	}
}

func checkFileInfos(t *testing.T, fi []os.FileInfo, expected []os.FileInfo) {
	if len(fi) != len(expected) {
		t.Errorf("len(fi)=%d, expected=%d", len(fi), len(expected))
	}
	l := len(fi)
	if len(expected) < l {
		l = len(expected)
	}
	for i := 0; i < l; i++ {
		checkFileInfo(t, fi[i], expected[i])
	}
}

func TestRecurReadDirFull(t *testing.T) {
	r := recurDirReader{
		dir:         "archive",
		readDirFunc: fakeReaderDirFunc(newFakeFS())}
	fis, err := r.recurReadDir()
	if err != nil {
		t.Fatalf("Unexpected error: %s\n", err)
	}

	checkFileInfos(t, fis, []os.FileInfo{
		&resultFileInfo{true, "archive/tar"},
		&resultFileInfo{false, "archive/tar/common.go"},
		&resultFileInfo{false, "archive/tar/example_test.go"},
		&resultFileInfo{false, "archive/tar/reader.go"},
		&resultFileInfo{false, "archive/tar/reader_test.go"},
		&resultFileInfo{true, "archive/tar/testdata"},
		&resultFileInfo{false, "archive/tar/testdata/gnu.tar"},
		&resultFileInfo{false, "archive/tar/testdata/pax.tar"},
		&resultFileInfo{false, "archive/tar/testdata/small.txt"},
		&resultFileInfo{false, "archive/tar/writer.go"},
		&resultFileInfo{false, "archive/tar/writer_test.go"},
		&resultFileInfo{true, "archive/zip"},
		&resultFileInfo{false, "archive/zip/example_test.go"},
		&resultFileInfo{false, "archive/zip/reader.go"},
		&resultFileInfo{false, "archive/zip/reader_test.go"},
		&resultFileInfo{false, "archive/zip/struct.go"},
		&resultFileInfo{true, "archive/zip/testdata"},
		&resultFileInfo{false, "archive/zip/testdata/crc32-not-streamed.zip"},
		&resultFileInfo{false, "archive/zip/testdata/dd.zip"},
		&resultFileInfo{false, "archive/zip/testdata/go-no-datadesc-sig.zip"},
		&resultFileInfo{false, "archive/zip/writer.go"},
		&resultFileInfo{false, "archive/zip/writer_test.go"},
		&resultFileInfo{false, "archive/zip/zip_test.go"},
	})
}

func TestRecurReadDirLimitCase1(t *testing.T) {
	r := recurDirReader{
		dir: "archive", maxEntries: 6,
		readDirFunc: fakeReaderDirFunc(newFakeFS())}
	fis, err := r.recurReadDir()
	if err != nil {
		t.Fatalf("Unexpected error: %s\n", err)
	}

	checkFileInfos(t, fis, []os.FileInfo{
		&resultFileInfo{true, "archive/tar"},
		&resultFileInfo{false, "archive/tar/common.go"},
		&resultFileInfo{false, "archive/tar/example_test.go"},
		&resultFileInfo{false, "archive/tar/reader.go"},
		&resultFileInfo{false, "archive/tar/reader_test.go"},
		&resultFileInfo{true, "archive/tar/testdata"},
	})
}

func TestRecurReadDirLimitCase2(t *testing.T) {
	r := recurDirReader{
		dir: "archive", maxEntries: 7,
		readDirFunc: fakeReaderDirFunc(newFakeFS())}
	fis, err := r.recurReadDir()
	if err != nil {
		t.Fatalf("Unexpected error: %s\n", err)
	}

	checkFileInfos(t, fis, []os.FileInfo{
		&resultFileInfo{true, "archive/tar"},
		&resultFileInfo{false, "archive/tar/common.go"},
		&resultFileInfo{false, "archive/tar/example_test.go"},
		&resultFileInfo{false, "archive/tar/reader.go"},
		&resultFileInfo{false, "archive/tar/reader_test.go"},
		&resultFileInfo{true, "archive/tar/testdata"},
		&resultFileInfo{false, "archive/tar/testdata/gnu.tar"},
	})
}

func TestRecurReadDirMarkerCase1(t *testing.T) {
	r := recurDirReader{
		dir: "archive", marker: "archive/tar/reader_test.go",
		readDirFunc: fakeReaderDirFunc(newFakeFS())}
	fis, err := r.recurReadDir()
	if err != nil {
		t.Fatalf("Unexpected error: %s\n", err)
	}

	checkFileInfos(t, fis, []os.FileInfo{
		&resultFileInfo{true, "archive/tar/testdata"},
		&resultFileInfo{false, "archive/tar/testdata/gnu.tar"},
		&resultFileInfo{false, "archive/tar/testdata/pax.tar"},
		&resultFileInfo{false, "archive/tar/testdata/small.txt"},
		&resultFileInfo{false, "archive/tar/writer.go"},
		&resultFileInfo{false, "archive/tar/writer_test.go"},
		&resultFileInfo{true, "archive/zip"},
		&resultFileInfo{false, "archive/zip/example_test.go"},
		&resultFileInfo{false, "archive/zip/reader.go"},
		&resultFileInfo{false, "archive/zip/reader_test.go"},
		&resultFileInfo{false, "archive/zip/struct.go"},
		&resultFileInfo{true, "archive/zip/testdata"},
		&resultFileInfo{false, "archive/zip/testdata/crc32-not-streamed.zip"},
		&resultFileInfo{false, "archive/zip/testdata/dd.zip"},
		&resultFileInfo{false, "archive/zip/testdata/go-no-datadesc-sig.zip"},
		&resultFileInfo{false, "archive/zip/writer.go"},
		&resultFileInfo{false, "archive/zip/writer_test.go"},
		&resultFileInfo{false, "archive/zip/zip_test.go"},
	})
}

func TestRecurReadDirMarkerCase2(t *testing.T) {
	r := recurDirReader{
		dir: "archive", marker: "archive/tar/reader.go",
		readDirFunc: fakeReaderDirFunc(newFakeFS())}
	fis, err := r.recurReadDir()
	if err != nil {
		t.Fatalf("Unexpected error: %s\n", err)
	}

	checkFileInfos(t, fis, []os.FileInfo{
		&resultFileInfo{false, "archive/tar/reader_test.go"},
		&resultFileInfo{true, "archive/tar/testdata"},
		&resultFileInfo{false, "archive/tar/testdata/gnu.tar"},
		&resultFileInfo{false, "archive/tar/testdata/pax.tar"},
		&resultFileInfo{false, "archive/tar/testdata/small.txt"},
		&resultFileInfo{false, "archive/tar/writer.go"},
		&resultFileInfo{false, "archive/tar/writer_test.go"},
		&resultFileInfo{true, "archive/zip"},
		&resultFileInfo{false, "archive/zip/example_test.go"},
		&resultFileInfo{false, "archive/zip/reader.go"},
		&resultFileInfo{false, "archive/zip/reader_test.go"},
		&resultFileInfo{false, "archive/zip/struct.go"},
		&resultFileInfo{true, "archive/zip/testdata"},
		&resultFileInfo{false, "archive/zip/testdata/crc32-not-streamed.zip"},
		&resultFileInfo{false, "archive/zip/testdata/dd.zip"},
		&resultFileInfo{false, "archive/zip/testdata/go-no-datadesc-sig.zip"},
		&resultFileInfo{false, "archive/zip/writer.go"},
		&resultFileInfo{false, "archive/zip/writer_test.go"},
		&resultFileInfo{false, "archive/zip/zip_test.go"},
	})
}

func TestRecurReadDirMarkerLimitCase1(t *testing.T) {
	r := recurDirReader{
		dir: "archive", marker: "archive/tar/reader.go",
		maxEntries: 4, readDirFunc: fakeReaderDirFunc(newFakeFS())}
	fis, err := r.recurReadDir()
	if err != nil {
		t.Fatalf("Unexpected error: %s\n", err)
	}

	checkFileInfos(t, fis, []os.FileInfo{
		&resultFileInfo{false, "archive/tar/reader_test.go"},
		&resultFileInfo{true, "archive/tar/testdata"},
		&resultFileInfo{false, "archive/tar/testdata/gnu.tar"},
		&resultFileInfo{false, "archive/tar/testdata/pax.tar"},
	})
}

func TestRecurReadDirMarkerLimitCase2(t *testing.T) {
	r := recurDirReader{
		dir: "archive", marker: "archive/tar/reader.go",
		maxEntries: 8, readDirFunc: fakeReaderDirFunc(newFakeFS())}
	fis, err := r.recurReadDir()
	if err != nil {
		t.Fatalf("Unexpected error: %s\n", err)
	}

	checkFileInfos(t, fis, []os.FileInfo{
		&resultFileInfo{false, "archive/tar/reader_test.go"},
		&resultFileInfo{true, "archive/tar/testdata"},
		&resultFileInfo{false, "archive/tar/testdata/gnu.tar"},
		&resultFileInfo{false, "archive/tar/testdata/pax.tar"},
		&resultFileInfo{false, "archive/tar/testdata/small.txt"},
		&resultFileInfo{false, "archive/tar/writer.go"},
		&resultFileInfo{false, "archive/tar/writer_test.go"},
		&resultFileInfo{true, "archive/zip"},
	})
}

func TestRecurReadDirMarkerLimitCase3(t *testing.T) {
	r := recurDirReader{
		dir: "archive", marker: "archive/tar/testdata",
		maxEntries: 6, readDirFunc: fakeReaderDirFunc(newFakeFS())}
	fis, err := r.recurReadDir()
	if err != nil {
		t.Fatalf("Unexpected error: %s\n", err)
	}

	checkFileInfos(t, fis, []os.FileInfo{
		&resultFileInfo{false, "archive/tar/testdata/gnu.tar"},
		&resultFileInfo{false, "archive/tar/testdata/pax.tar"},
		&resultFileInfo{false, "archive/tar/testdata/small.txt"},
		&resultFileInfo{false, "archive/tar/writer.go"},
		&resultFileInfo{false, "archive/tar/writer_test.go"},
		&resultFileInfo{true, "archive/zip"},
	})
}

func TestRecurReadDirMarkerLimitCase4(t *testing.T) {
	r := recurDirReader{
		dir: "archive", marker: "archive/tar/testdata",
		maxEntries: 7, readDirFunc: fakeReaderDirFunc(newFakeFS())}
	fis, err := r.recurReadDir()
	if err != nil {
		t.Fatalf("Unexpected error: %s\n", err)
	}

	checkFileInfos(t, fis, []os.FileInfo{
		&resultFileInfo{false, "archive/tar/testdata/gnu.tar"},
		&resultFileInfo{false, "archive/tar/testdata/pax.tar"},
		&resultFileInfo{false, "archive/tar/testdata/small.txt"},
		&resultFileInfo{false, "archive/tar/writer.go"},
		&resultFileInfo{false, "archive/tar/writer_test.go"},
		&resultFileInfo{true, "archive/zip"},
		&resultFileInfo{false, "archive/zip/example_test.go"},
	})
}

func TestRecurReadDirMarkerLimitMatcher(t *testing.T) {
	matcher, err := NewMatcher(nil, []string{"**/testdata/**"})
	if err != nil {
		t.Fatalf("Unexpected error: %s\n", err)
	}
	r := recurDirReader{
		dir: "archive", marker: "archive/tar/reader.go",
		maxEntries: 8, matcher: matcher,
		readDirFunc: fakeReaderDirFunc(newFakeFS())}
	fis, err := r.recurReadDir()
	if err != nil {
		t.Fatalf("Unexpected error: %s\n", err)
	}

	checkFileInfos(t, fis, []os.FileInfo{
		&resultFileInfo{false, "archive/tar/reader_test.go"},
		&resultFileInfo{false, "archive/tar/writer.go"},
		&resultFileInfo{false, "archive/tar/writer_test.go"},
		&resultFileInfo{true, "archive/zip"},
		&resultFileInfo{false, "archive/zip/example_test.go"},
		&resultFileInfo{false, "archive/zip/reader.go"},
		&resultFileInfo{false, "archive/zip/reader_test.go"},
		&resultFileInfo{false, "archive/zip/struct.go"},
	})
}
