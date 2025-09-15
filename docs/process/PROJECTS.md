# Project Sync Configuration and Validation

This document describes the configuration and validation process for syncing issues and pull requests to the Game Roadmap project.

## Configuration Requirements

### Repository Variables
- **PROJECT_URL**: Set to `https://github.com/users/AstroSteveo/projects/2`
  - This is configured at the repository level in Settings > Secrets and variables > Actions > Variables
  - Used by the project-sync workflow to identify the target project

### Repository Secrets
- **PROJECTS_TOKEN**: Personal Access Token with required scopes
  - Required scopes: `project`, `repo`
  - This is configured at the repository level in Settings > Secrets and variables > Actions > Secrets
  - Used by the project-sync workflow for authentication

## Workflow Configuration

The project-sync workflow (`.github/workflows/project-sync.yml`) is configured to:
- Trigger on issue and pull request events: opened, labeled, closed, reopened
- Add items with labels: `story`, `bug`, `task`
- Use the `actions/add-to-project@v0.6.0` action
- Skip execution for dependabot and when PROJECT_URL is not configured

## End-to-End Validation Checklist

Use this checklist to validate the project sync functionality:

### Prerequisites
- [ ] Repository variable `PROJECT_URL` is set to `https://github.com/users/AstroSteveo/projects/2`
- [ ] Repository secret `PROJECTS_TOKEN` exists with scopes: `project`, `repo`
- [ ] Game Roadmap project exists at the specified URL
- [ ] Token has access to the project

### Test Story Issue
- [ ] Create a new issue using the "Story" template
- [ ] Verify issue is automatically labeled with `story`
- [ ] Check that project-sync workflow runs successfully
- [ ] Confirm issue appears in the Game Roadmap project
- [ ] Verify issue status/metadata is correctly synced

### Test Task Issue  
- [ ] Create a new issue using the "Task" template
- [ ] Verify issue is automatically labeled with `task`
- [ ] Check that project-sync workflow runs successfully
- [ ] Confirm issue appears in the Game Roadmap project

### Test Bug Issue
- [ ] Create a new issue using the "Bug" template  
- [ ] Verify issue is automatically labeled with `bug`
- [ ] Check that project-sync workflow runs successfully
- [ ] Confirm issue appears in the Game Roadmap project

### Test Pull Request Sync
- [ ] Create a pull request and add one of the labels: `story`, `bug`, or `task`
- [ ] Check that project-sync workflow runs successfully
- [ ] Confirm pull request appears in the Game Roadmap project

### Test Workflow Conditions
- [ ] Verify workflow skips execution when PROJECT_URL is empty
- [ ] Verify workflow skips execution for dependabot PRs
- [ ] Test workflow with unlabeled issues (should not sync)

### Troubleshooting
If sync fails, check:
1. **Workflow Logs**: Go to Actions tab and check project-sync workflow logs
2. **Token Permissions**: Ensure PROJECTS_TOKEN has `project` and `repo` scopes
3. **Project Access**: Verify token has access to the target project
4. **Label Matching**: Ensure issue/PR has one of: `story`, `bug`, `task` labels

## Manual Sync Process

If automatic sync fails, items can be manually added to the project:
1. Navigate to the [Game Roadmap project](https://github.com/users/AstroSteveo/projects/2)
2. Click "Add item" 
3. Search for and select the issue/PR
4. Configure appropriate status and metadata

## Related Files

- `.github/workflows/project-sync.yml` - Main workflow configuration
- `.github/ISSUE_TEMPLATE/story.yml` - Story issue template with `story` label
- `.github/ISSUE_TEMPLATE/task.yml` - Task issue template with `task` label  
- `.github/ISSUE_TEMPLATE/bug.yml` - Bug issue template with `bug` label