---
name: release
description: Commit, push, tag, and release timesheetz
disable-model-invocation: true
---

# Release Skill

Create a new release for timesheetz. Usage: `/release [major|minor|patch]`

Default: `minor` version bump.

## Steps

1. **Check status**: Run `git status` and `git diff --stat` to see uncommitted changes
2. **Abort if clean**: If no changes, inform user and stop
3. **Create commit**:
   - Write a descriptive commit message summarizing the changes
   - Always end with: `Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>`
   - Use HEREDOC format for the commit message
4. **Handle remote changes**: Run `git pull --rebase origin main` before pushing
5. **Push**: Run `git push origin main`
6. **Determine version**:
   - Get latest tag: `git tag --sort=-v:refname | head -1`
   - Parse current version (e.g., v1.29.0)
   - Bump based on argument:
     - `patch`: v1.29.0 → v1.29.1
     - `minor` (default): v1.29.0 → v1.30.0
     - `major`: v1.29.0 → v2.0.0
7. **Create tag**: `git tag <new-version>`
8. **Push tag**: `git push origin <new-version>`
9. **Create release**: Check if GitHub Action created it, otherwise create with `gh release create`
10. **Report**: Show the release URL to the user

## Argument Handling

- `/release` → minor bump
- `/release patch` → patch bump
- `/release minor` → minor bump
- `/release major` → major bump

## Commit Message Format

```
<Summary of changes in imperative mood>

<Optional bullet points for details>

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
```

## Safety

- Never force push
- Never skip hooks
- Always pull --rebase before pushing
- Wait for user confirmation if there are merge conflicts
