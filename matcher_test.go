package paths_test

import (
	. "github.com/hnakamur/paths"
	"testing"
)

type testCase struct {
	path    string
	matches bool
}

func Example() error {
	matcher, err := NewMatcher(
		[]string{"docs/", "src/"},
		append(DefaultExcludes, "**/*.o", "**/*.a"))
	if err != nil {
		return err
	}

	if matcher.Match("foo/bar.o") {
		println("matched")
	} else {
		println("not matched")
	}
	return nil
}

func testMatcher(t *testing.T, includes, excludes []string, cases []testCase) {
	matcher, err := NewMatcher(includes, excludes)
	if err != nil {
		t.Fatal(err)
	}

	for _, c := range cases {
		actual := matcher.Match(c.path)
		if actual != c.matches {
			t.Errorf("path:%s\texpected:%v\tactual:%v", c.path, c.matches, actual)
		}
	}
}

func TestOnlyIncludes(t *testing.T) {
	testMatcher(t,
		[]string{
			"**/*.go",
			"*.txt",
			"bar/",
			"baz/**/*.?",
		},
		nil,
		[]testCase{
			{"foo.go", true},
			{"bar.go", true},
			{"foo/bar.go", true},
			{"readme.txt", true},
			{"foo/readme.txt", false},
			{"readme.html", false},
			{"bar/readme.html", true},
			{"bar", true},
			{"baz/foo.c", true},
			{"baz/foo/bar/foo.c", true},
			{"baz/foo/bar/foo.txt", false},
		})
}

func TestAntExample1(t *testing.T) {
	testMatcher(t,
		[]string{
			"**/CVS/*",
		},
		nil,
		[]testCase{
			{"CVS/Repository", true},
			{"org/apache/CVS/Entries", true},
			{"org/apache/jakarta/tools/ant/CVS/Entries", true},
			{"org/apache/CVS/foo/bar/Entries", false},
		})
}

func TestAntExample2(t *testing.T) {
	testMatcher(t,
		[]string{
			"org/apache/jakarta/**",
		},
		nil,
		[]testCase{
			{"org/apache/jakarta/tools/ant/docs/index.html", true},
			{"org/apache/jakarta/test.xml", true},
			{"org/apache/xyz.java", false},
		})
}

func TestAntExample3(t *testing.T) {
	testMatcher(t,
		[]string{
			"org/apache/**/CVS/*",
		},
		nil,
		[]testCase{
			{"org/apache/CVS/Entries", true},
			{"org/apache/jakarta/tools/ant/CVS/Entries", true},
			{"org/apache/CVS/foo/bar/Entries", false},
		})
}

func TestAntExample4(t *testing.T) {
	testMatcher(t,
		[]string{
			"**/test/**",
		},
		nil,
		[]testCase{
			{"org/apache/test/CVS/Entries", true},
			{"org/apache/test/CVS", true},
			{"org/apache/test", true},
			{"test", true},
			{"test/CVS", true},
			{"test/CVS/Entires", true},
			{"org/apache/CVS/Entries", false},
		})
}

func TestDefaultExcludes(t *testing.T) {
	testMatcher(t,
		nil,
		DefaultExcludes,
		[]testCase{
			{"foo.go", true},
			{"foo.go~", false},
			{"foo/bar.go~", false},
			{"foo/.svn", false},
			{"foo/.svn/format", false},
		})
}

func TestOnlyExcludes(t *testing.T) {
	testMatcher(t,
		nil,
		append(DefaultExcludes, "**/*.o", "**/*.a"),
		[]testCase{
			{"foo.c", true},
			{"foo.o", false},
			{"lib/foo.a", false},
		})
}

func TestIncludesAndExcludes(t *testing.T) {
	testMatcher(t,
		[]string{
			"*.md",
			"src/",
		},
		append(DefaultExcludes, "**/*.o", "**/*.a"),
		[]testCase{
			{"README.md", true},
			{"tmp/foo.md", false},
			{"src/foo.c", true},
			{"src/foo.o", false},
			{"lib/foo.a", false},
		})
}

func TestEscapeMeta(t *testing.T) {
	testMatcher(t, []string{"a|b"}, nil, []testCase{{"a|b", true}})
	testMatcher(t, []string{"a+b"}, nil, []testCase{{"a+b", true}})
	testMatcher(t, []string{"a[b"}, nil, []testCase{{"a[b", true}})
	testMatcher(t, []string{"a{b"}, nil, []testCase{{"a{b", true}})
	testMatcher(t, []string{"a(b"}, nil, []testCase{{"a(b", true}})
	testMatcher(t, []string{"a^b"}, nil, []testCase{{"a^b", true}})
	testMatcher(t, []string{"a$b"}, nil, []testCase{{"a$b", true}})
	testMatcher(t, []string{`a\b`}, nil, []testCase{{`a\b`, true}})
}
