package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"unicode"
)

var allowedTypes = []string{
	"feat", "fix", "docs", "style", "refactor",
	"perf", "test", "build", "ci", "chore", "revert",
}

var nonImperative = map[string]string{
	"added":     "add",
	"adds":      "add",
	"adding":    "add",
	"fixed":     "fix",
	"fixes":     "fix",
	"fixing":    "fix",
	"updated":   "update",
	"updates":   "update",
	"updating":  "update",
	"removed":   "remove",
	"removes":   "remove",
	"removing":  "remove",
	"changed":   "change",
	"changes":   "change",
	"changing":  "change",
	"created":   "create",
	"creates":   "create",
	"creating":  "create",
	"refactored": "refactor",
	"refactors":  "refactor",
	"refactoring": "refactor",
}

var headerRE = regexp.MustCompile(`^(?P<type>[a-z]+)(?:\((?P<scope>[a-z0-9_/-]+)\))?(?P<bang>!?): (?P<subject>.+)$`)

type finding struct {
	rule    string
	message string
	fix     string
}

func main() {
	fixMode := flag.Bool("fix", false, "suggest fixes for detected issues")
	maxLen := flag.Int("max-subject", 72, "maximum subject length in characters")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: commitlint [--fix] [--max-subject N] [<msg-file>]")
		fmt.Fprintln(os.Stderr, "if no file is given, reads git log -1 --format=%B from the current repo")
		flag.PrintDefaults()
	}
	flag.Parse()

	msg, err := readMessage(flag.Args())
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(2)
	}

	findings := lint(msg, *maxLen)
	if len(findings) == 0 {
		fmt.Println("ok: commit message passes all checks")
		return
	}
	for _, f := range findings {
		fmt.Printf("[%s] %s\n", f.rule, f.message)
		if *fixMode && f.fix != "" {
			fmt.Printf("        fix: %s\n", f.fix)
		}
	}
	os.Exit(1)
}

func readMessage(args []string) (string, error) {
	if len(args) >= 1 {
		b, err := os.ReadFile(args[0])
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
	cmd := exec.Command("git", "log", "-1", "--format=%B")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git log failed: %w", err)
	}
	return string(out), nil
}

func lint(raw string, maxLen int) []finding {
	var findings []finding
	msg := strings.TrimRight(raw, "\n")
	if strings.TrimSpace(msg) == "" {
		return []finding{{rule: "empty", message: "commit message is empty"}}
	}
	lines := splitLines(msg)
	header := lines[0]

	if len(header) > maxLen {
		findings = append(findings, finding{
			rule:    "subject-max-length",
			message: fmt.Sprintf("header is %d chars (max %d)", len(header), maxLen),
			fix:     "shorten the header or move detail to the body",
		})
	}

	m := headerRE.FindStringSubmatch(header)
	if m == nil {
		findings = append(findings, finding{
			rule:    "header-format",
			message: "header does not match: <type>(<scope>)?!?: <subject>",
			fix:     "example: feat(auth): add OAuth login flow",
		})
		return findings
	}
	typeName := m[headerRE.SubexpIndex("type")]
	subject := m[headerRE.SubexpIndex("subject")]

	if !contains(allowedTypes, typeName) {
		findings = append(findings, finding{
			rule:    "type-enum",
			message: fmt.Sprintf("type %q not in {%s}", typeName, strings.Join(allowedTypes, ", ")),
			fix:     "use one of the allowed types listed above",
		})
	}

	if len(subject) == 0 {
		findings = append(findings, finding{
			rule:    "subject-empty",
			message: "subject is empty",
			fix:     "describe the change in a short imperative sentence",
		})
	} else {
		firstRune := []rune(subject)[0]
		if unicode.IsUpper(firstRune) {
			findings = append(findings, finding{
				rule:    "subject-lower-case",
				message: "subject should start with a lower-case letter",
				fix:     lowerFirst(subject),
			})
		}
		if strings.HasSuffix(subject, ".") {
			findings = append(findings, finding{
				rule:    "subject-no-period",
				message: "subject must not end with a period",
				fix:     strings.TrimRight(subject, "."),
			})
		}
		firstWord := strings.ToLower(strings.Fields(subject)[0])
		if suggestion, ok := nonImperative[firstWord]; ok {
			findings = append(findings, finding{
				rule:    "imperative-mood",
				message: fmt.Sprintf("subject should use imperative mood (got %q)", firstWord),
				fix:     replaceFirstWord(subject, suggestion),
			})
		}
	}

	if len(lines) >= 2 && strings.TrimSpace(lines[1]) != "" {
		findings = append(findings, finding{
			rule:    "blank-line-after-header",
			message: "second line must be blank",
			fix:     "insert an empty line between header and body",
		})
	}

	if len(lines) >= 3 {
		for i, l := range lines[2:] {
			if len(l) > 100 {
				findings = append(findings, finding{
					rule:    "body-line-length",
					message: fmt.Sprintf("body line %d is %d chars (max 100)", i+3, len(l)),
					fix:     "hard-wrap the body at 100 columns",
				})
				break
			}
		}
	}

	return findings
}

func splitLines(s string) []string {
	var out []string
	sc := bufio.NewScanner(strings.NewReader(s))
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		out = append(out, sc.Text())
	}
	return out
}

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

func lowerFirst(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToLower(r[0])
	return string(r)
}

func replaceFirstWord(s, w string) string {
	fields := strings.SplitN(s, " ", 2)
	if len(fields) == 1 {
		return w
	}
	return w + " " + fields[1]
}
