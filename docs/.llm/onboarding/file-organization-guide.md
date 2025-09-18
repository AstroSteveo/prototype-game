# File Organization and Cleanup Guide for AI Agents

This guide ensures AI agents maintain a tidy, clear, and contradiction-free documentation structure by following best practices for file management and content organization.

## Core Principles

### 1. Always Clean Before Creating
- **Check for existing content**: Before creating new files, thoroughly search for existing files with similar content
- **Reuse over recreation**: If a file exists with appropriate content, extend or modify it rather than creating a duplicate
- **Delete obsolete files**: Remove files that are no longer needed or have been superseded

### 2. File Naming Conventions
- Use descriptive, specific names that clearly indicate the file's purpose
- Avoid creating multiple files with similar names (e.g., `guide.md`, `guide-new.md`, `guide-v2.md`)
- Follow existing naming patterns in the directory

### 3. Content Consolidation
- Merge related content into single files when appropriate
- Avoid spreading similar information across multiple files
- Use sections and headings to organize content within files

## File Management Workflow

### Before Creating Any New File

1. **Search for existing content**:
   ```bash
   # Search for files with similar names
   find . -name "*keyword*" -type f
   
   # Search for similar content
   grep -r "relevant terms" docs/
   ```

2. **Evaluate existing files**:
   - Can the content be added to an existing file?
   - Would merging improve organization?
   - Is the existing file outdated and should be replaced?

3. **Choose the appropriate action**:
   - **Extend**: Add content to existing file
   - **Replace**: Update existing file with new content
   - **Merge**: Combine multiple files into one
   - **Create**: Only if no suitable existing file exists

### File Cleanup Checklist

- [ ] **Remove duplicate content**: Check for information that exists in multiple places
- [ ] **Delete outdated files**: Remove files that reference non-existent paths or outdated information
- [ ] **Fix broken references**: Update all links pointing to moved or deleted files
- [ ] **Consolidate similar files**: Merge files with overlapping purposes
- [ ] **Update cross-references**: Ensure all internal links are current and functional

## Directory-Specific Guidelines

### `/docs/.llm/onboarding/`
- Keep onboarding files focused and non-overlapping
- Merge guides with similar purposes
- Ensure each file has a distinct, clear role

### `/docs/process/`
- Avoid creating multiple session templates for similar purposes
- Consolidate related process documentation
- Remove outdated process files

### `/scripts/`
- Remove unused or broken scripts
- Consolidate scripts with similar functionality
- Update script references in documentation

### Root-level files
- Keep only essential files at repository root
- Move specialized documentation to appropriate subdirectories
- Avoid proliferation of README-style files

## Content Validation

### Before Committing Changes

1. **Run validation scripts**:
   ```bash
   ./scripts/agents/validate-onboarding.sh
   ```

2. **Check for broken references**:
   ```bash
   # Custom command to find broken internal links
   grep -r "\[\[.*\]\]" docs/ || true
   grep -r "\[.*\](docs/" . | grep -v ".git"
   ```

3. **Verify no duplicate content**:
   - Manually review related files for overlap
   - Ensure each piece of information has a single source of truth

## Automation Guidelines

### For AI Agents

- **Always run cleanup before creation**: Use the search and evaluation workflow
- **Prefer modification over creation**: Default to extending existing files
- **Clean up after changes**: Remove any temporary or obsolete files created during the process
- **Validate systematically**: Use validation scripts to ensure changes don't break references

### Automated Checks

The validation framework should check for:
- Broken internal references
- Duplicate file content
- Outdated file references
- Unused files in the repository

## Examples of Good Practices

### ✅ Good: Consolidating Similar Files
```
Before:
- quick-start.md
- getting-started.md 
- startup-guide.md

After:
- quick-start.md (consolidated content from all three)
```

### ✅ Good: Fixing References During Cleanup
```
Before:
- Reference to docs/design/GDD.md (non-existent)

After:
- Reference to docs/product/vision/game-design-document.md (actual file)
```

### ✅ Good: Extending Existing Files
```
Before:
- Creating new file for agent validation

After:
- Adding section to existing agent-validation-checklist.md
```

### ❌ Bad: Creating Similar Files
```
- agent-guide.md
- agent-instructions.md
- agent-manual.md
(All covering similar ground)
```

## Maintenance Schedule

### Regular Cleanup Tasks

1. **Weekly**: Check for and remove temporary files
2. **Per PR**: Validate all references are functional
3. **Monthly**: Review for duplicate content and consolidation opportunities
4. **Quarterly**: Full repository organization review

### Validation Integration

- All validation scripts should check for file organization issues
- CI should prevent commits with broken references
- Regular automated checks for duplicate content

## Integration with Existing Framework

This guide complements:
- `docs/.llm/onboarding/contribution-checklist.md` - Add file organization steps
- `scripts/agents/validate-onboarding.sh` - Enhance with duplicate content checks
- `docs/.llm/AGENTS.md` - Reference this guide for file management expectations

## Summary

By following these guidelines, AI agents will:
- Maintain a clean, organized repository structure
- Avoid creating duplicate or conflicting documentation
- Ensure all references remain functional
- Keep the documentation ecosystem tidy and contradiction-free

Remember: **Always clean before creating, always reuse before recreating, always validate before committing.**