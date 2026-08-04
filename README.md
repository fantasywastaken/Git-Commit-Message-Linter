# Git-Commit-Message-Linter

Conventional Commits validator that lints a commit message file or the most recent commit and can suggest fixes.

### ⚙️ How It Works

The linter reads the commit message either from a file argument (the standard `commit-msg` hook contract) or by shelling out to `git log -1 --format=%B`. It matches the header against a Conventional Commits regexp `type(scope)?!?: subject`, verifies the type against a fixed vocabulary (`feat`, `fix`, `docs`, ...), and enforces subject rules: length, lower-case first letter, no trailing period, and imperative mood. It also checks that the header is followed by a blank line and that body lines do not exceed 100 columns. Findings are printed with a rule id; `--fix` adds a suggested rewrite line beneath each finding. Any finding exits `1`.

## 📁 Setup

### Requirements
- Go 1.22 or newer
- `git` on the `PATH` (only needed when no message file is passed)

### Installation
```bash
git clone https://github.com/fantasywastaken/Git-Commit-Message-Linter.git
cd Git-Commit-Message-Linter
go build -o commitlint .
```

### 🚀 Usage
```bash
commitlint .git/COMMIT_EDITMSG
commitlint --fix .git/COMMIT_EDITMSG
commitlint --max-subject 80
commitlint
```

Add it as a `commit-msg` hook:
```bash
echo 'exec commitlint "$1"' > .git/hooks/commit-msg
chmod +x .git/hooks/commit-msg
```

### ✨ Features
- Validates Conventional Commits header: type, optional scope, optional `!` breaking marker, subject
- Configurable subject length (default 72)
- Detects non-imperative first words (`added`, `fixes`, `updating`, ...) and suggests the imperative
- Enforces lower-case subject start and no trailing period
- Requires a blank line between header and body
- Warns on body lines longer than 100 columns
- `--fix` prints suggested rewrites next to each finding
- Non-zero exit code, so it drops straight into `commit-msg` hooks and CI
