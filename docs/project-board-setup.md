# Project Board Setup Guide

This document provides step-by-step instructions for setting up the GitHub Project Board automation with the prototype-game repository.

## Overview

The project board automation provides seamless integration between GitHub issues, pull requests, and project management using:

- Automated project item addition for new issues and PRs
- Status synchronization based on labels and issue/PR states
- Milestone and estimate tracking
- Sprint/iteration management

## Prerequisites

- Repository admin access to AstroSteveo/prototype-game
- Access to GitHub Projects at https://github.com/users/AstroSteveo/projects/2
- Fine-grained Personal Access Token with appropriate permissions

## Setup Instructions

### 1. Repository Configuration

#### Set Repository Variables
Navigate to repository Settings → Secrets and variables → Actions → Variables and add:

- **Variable Name**: `PROJECT_URL`
- **Value**: `https://github.com/users/AstroSteveo/projects/2`

#### Set Repository Secrets
Navigate to repository Settings → Secrets and variables → Actions → Secrets and add:

- **Secret Name**: `PROJECTS_TOKEN`
- **Value**: [Your fine-grained Personal Access Token]

**Token Requirements:**
- Scope: Repository and Project access
- Permissions:
  - Repository: Read access to metadata, issues, pull requests
  - Projects: Write access to projects

### 2. Project Field Configuration

Visit: https://github.com/users/AstroSteveo/projects/2/settings/fields

Create the following custom fields:

#### Status Field (Single-select)
- **Field Name**: Status
- **Options**:
  - Backlog
  - Ready
  - In Progress
  - In Review
  - Blocked
  - Done

#### Estimate Field (Number)
- **Field Name**: Estimate
- **Description**: Story points or effort estimate
- **Number format**: Integer

#### Milestone Field (Milestone)
- **Field Name**: Milestone
- **Links to**: Repository milestones

#### Sprint Field (Iteration)
- **Field Name**: Sprint
- **Duration**: 2 weeks (adjust as needed)
- **Start date**: Configure according to team schedule

### 3. Project Automation Rules

Visit: https://github.com/users/AstroSteveo/projects/2/workflows

Configure the following automation rules:

#### Item Added Automation
- **Trigger**: Item added to project
- **Action**: Set Status to "Backlog"

#### Assignment Automation
- **Trigger**: Issue assigned
- **Action**: Set Status to "In Progress"

#### Issue Closed Automation
- **Trigger**: Issue closed
- **Action**: Set Status to "Done"

#### PR Merged Automation
- **Trigger**: Pull request merged
- **Action**: Set Status to "Done"

#### Auto-archive Automation
- **Trigger**: Item in "Done" status for 14 days
- **Action**: Archive item

### 4. Label-based Status Transitions

The workflow automatically updates status based on these labels:

| Label | New Status |
|-------|------------|
| `ready` | Ready |
| `blocked` | Blocked |
| `in-progress` | In Progress |
| `in-review` | In Review |

### 5. Validation Steps

#### Test Issue Creation
1. Create a new issue using the Task template
2. Title: "task: Test automation [2]"
3. Add label: `task`
4. Verify:
   - Issue appears in project
   - Status is set to "Backlog"
   - Estimate is extracted from title (2)

#### Test Status Transitions
1. Add `ready` label → Status should become "Ready"
2. Add `blocked` label → Status should become "Blocked"
3. Remove `blocked`, add `in-progress` → Status should become "In Progress"
4. Close issue → Status should become "Done"

#### Test PR Workflow
1. Open draft PR → Status: "In Progress"
2. Mark ready for review → Status: "In Review"
3. Merge PR → Status: "Done"

## Troubleshooting

### Common Issues

#### Workflow Not Triggering
- Verify `PROJECTS_TOKEN` secret is set correctly
- Check token permissions include project write access
- Ensure `PROJECT_URL` matches exact project URL

#### Status Not Updating
- Verify project fields exist with correct names
- Check project automation rules are enabled
- Review workflow logs in Actions tab

#### Items Not Being Added
- Confirm `actions/add-to-project` action has latest version
- Verify project URL format is correct
- Check repository variable configuration

### Workflow Logs

Monitor automation in the Actions tab:
- Navigate to repository → Actions
- Look for "Project Sync" workflow runs
- Check logs for any error messages

## Maintenance

### Regular Tasks
- Review and update automation rules as workflow evolves
- Monitor token expiration and refresh as needed
- Adjust sprint duration and field options based on team needs

### Sprint Management
- Create new sprints/iterations in project settings
- Archive completed sprints to maintain organization
- Update sprint assignments during sprint planning

## Support

For issues with this setup:
1. Check workflow logs in GitHub Actions
2. Verify all configuration steps were completed
3. Review GitHub Projects documentation
4. Contact repository administrators for access issues

---

Last updated: September 2025