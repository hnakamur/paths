package paths

import (
	"regexp"
	"strings"
)

type Matcher interface {
	Match(path string) bool
}

type matcherRegexp struct {
	include *regexp.Regexp
	exclude *regexp.Regexp
}

// NewMatcher returns a new Matcher for include and exclude glob patterns.
// When both inclusion and exclusion are used, only paths that match at
// least one of the include patterns and don't match any of the exclude
// patterns are used. If no include pattern is given, all paths are assumed
// to match the include pattern (with the possible exception of the excludes).
//
// Patterns are used for the inclusion and exclusion of paths.
// '*' matches zero or more characters, '?' matches one character.
// '*' and '?' does not match a directory separator '/'.
//
// '**/' at the beginning or '/**/' in the middle matches zero or more
// direcotries. '/' or '/**' at the end matches any files or directories in
// the subdirectories.
func NewMatcher(includes, excludes []string) (Matcher, error) {
	include, err := convertGlobs(includes)
	if err != nil {
		return nil, err
	}

	exclude, err := convertGlobs(excludes)
	if err != nil {
		return nil, err
	}

	return &matcherRegexp{include, exclude}, nil
}

func (m *matcherRegexp) Match(path string) bool {
	return (m.include == nil || m.include.MatchString(path)) &&
		(m.exclude == nil || !m.exclude.MatchString(path))
}

var DefaultExcludes = []string{
	"**/*~",
	"**/#*#",
	"**/.#*",
	"**/%*%",
	"**/._*",
	"**/CVS/**",
	"**/.cvsignore",
	"**/SCCS/**",
	"**/vssver.scc",
	"**/.svn/**",
	"**/.DS_Store",
	"**/.git/**",
	"**/.gitattributes",
	"**/.gitignore",
	"**/.gitmodules",
	"**/.hg/**",
	"**/.hgignore",
	"**/.hgsub",
	"**/.hgsubstate",
	"**/.hgtags",
	"**/.bzr/**",
	"**/.bzrignore",
}

func convertGlobs(patterns []string) (*regexp.Regexp, error) {
	if len(patterns) == 0 {
		return nil, nil
	}

	exprs := make([]string, len(patterns))
	for i, pattern := range patterns {
		var prefix, suffix string

		if strings.HasPrefix(pattern, "**/") {
			prefix = "(.+/)?"
			pattern = pattern[len("**/"):]
		}

		if strings.HasSuffix(pattern, "/") {
			suffix = "(/.+)?"
			pattern = pattern[:len(pattern)-len("/")]
		} else if strings.HasSuffix(pattern, "/**") {
			suffix = "(/.+)?"
			pattern = pattern[:len(pattern)-len("/**")]
		}

		exprs[i] = prefix + replacer.Replace(pattern) + suffix
	}
	return regexp.Compile(`\A(` + strings.Join(exprs, "|") + `)\z`)
}

var replacer = strings.NewReplacer(
	"/**/", "/(.+/)?",
	"*", "[^/]*",
	"?", "[^/]",
	".", `\.`,
	"|", `\|`,
	"+", `\+`,
	"[", `\[`,
	"(", `\(`,
	"{", `\{`,
	"^", `\^`,
	"$", `\$`,
	`\`, `\\`,
)
