# Project Board Automation Test

This is a test issue to validate the project board automation functionality.

## Test Case Details
- **Title Pattern**: This issue title contains `[2]` to test estimate parsing
- **Expected Automation**:
  - ✅ Added to project (has `task` label)
  - ✅ Status set to Backlog (new issue)
  - ✅ Estimate set to 2 (from `[2]` in title)
  - ✅ Sprint set to current iteration (if configured)

## Manual Testing Steps
1. Create this issue with title ending in `[2]`
2. Add `task` label
3. Verify project automation:
   - Issue appears in project board
   - Status = Backlog
   - Estimate = 2
4. Test status transitions:
   - Add `ready` label → Status becomes Ready
   - Add `blocked` label → Status becomes Blocked
   - Remove labels, close issue → Status becomes Done

## Validation Checklist
- [ ] Issue added to project automatically
- [ ] Status field set to Backlog
- [ ] Estimate field set to 2
- [ ] Sprint field set (if iterations configured)
- [ ] Status transitions work with labels
- [ ] Closing issue sets Status to Done