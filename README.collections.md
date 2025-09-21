# 📦 Collections

Curated collections of related prompts, instructions, and chat modes organized around specific themes, workflows, or use cases.

## Effective State and Precedence Rules

When using collections alongside individual item settings, the effective state of each item follows these precedence rules:

1. **Explicit overrides take precedence**: If an item has an explicit `true` or `false` setting in your config, that always takes priority over collections
2. **Collections provide defaults**: If an item has no explicit setting (undefined), it inherits from enabled collections that contain it
3. **Union behavior**: An item is enabled if it's in ANY enabled collection OR explicitly enabled
4. **Shared items remain enabled**: Disabling one collection won't disable shared items if they're still required by other enabled collections

### Examples

**Explicit override wins:**
```yaml
collections:
  testing-automation: true
prompts:
  playwright-generate-test: false  # Stays disabled despite collection being enabled
```

**Inheritance from collections:**
```yaml  
collections:
  testing-automation: true
prompts:
  # playwright-generate-test: not specified - inherits from collection (enabled)
```

**Shared items protection:**
```yaml
collections:
  frontend-web-dev: true      # Contains playwright-generate-test
  testing-automation: false  # Also contains playwright-generate-test
prompts:
  # playwright-generate-test: enabled via frontend-web-dev
```

When you toggle collections, the CLI shows delta summaries:
- 📈 Items newly enabled by the collection
- 📉 Items disabled by disabling the collection  
- 🚫 Items blocked by explicit overrides

### How to Use Collections

**Browse Collections:**
- Explore themed collections that group related customizations
- Each collection includes prompts, instructions, and chat modes for specific workflows
- Collections make it easy to adopt comprehensive toolkits for particular scenarios

**Install Items:**
- Click install buttons for individual items within collections
- Or browse to the individual files to copy content manually
- Collections help you discover related customizations you might have missed

| Name | Description | Items | Tags |
| ---- | ----------- | ----- | ---- |
| [Azure & Cloud Development](collections/azure-cloud-development.md) | Comprehensive Azure cloud development tools including Infrastructure as Code, serverless functions, architecture patterns, and cost optimization for building scalable cloud applications. | 15 items | azure, cloud, infrastructure, bicep, terraform, serverless, architecture, devops |
| [C# .NET Development](collections/csharp-dotnet-development.md) | Essential prompts, instructions, and chat modes for C# and .NET development including testing, documentation, and best practices. | 7 items | csharp, dotnet, aspnet, testing |
| [Database & Data Management](collections/database-data-management.md) | Database administration, SQL optimization, and data management tools for PostgreSQL, SQL Server, and general database development best practices. | 8 items | database, sql, postgresql, sql-server, dba, optimization, queries, data-management |
| [DevOps On-Call](collections/devops-oncall.md) | A focused set of prompts, instructions, and a chat mode to help triage incidents and respond quickly with DevOps tools and Azure resources. | 5 items | devops, incident-response, oncall, azure |
| [Frontend Web Development](collections/frontend-web-dev.md) | Essential prompts, instructions, and chat modes for modern frontend web development including React, Angular, Vue, TypeScript, and CSS frameworks. | 11 items | frontend, web, react, typescript, javascript, css, html, angular, vue |
| [Project Planning & Management](collections/project-planning.md) | Tools and guidance for software project planning, feature breakdown, epic management, implementation planning, and task organization for development teams. | 17 items | planning, project-management, epic, feature, implementation, task, architecture, technical-spike |
| [Security & Code Quality](collections/security-best-practices.md) | Security frameworks, accessibility guidelines, performance optimization, and code quality best practices for building secure, maintainable, and high-performance applications. | 6 items | security, accessibility, performance, code-quality, owasp, a11y, optimization, best-practices |
| [Technical Spike](collections/technical-spike.md) | Tools for creation, management and research of technical spikes to reduce unknowns and assumptions before proceeding to specification and implementation of solutions. | 2 items | technical-spike, assumption-testing, validation, research |
| [Testing & Test Automation](collections/testing-automation.md) | Comprehensive collection for writing tests, test automation, and test-driven development including unit tests, integration tests, and end-to-end testing strategies. | 11 items | testing, tdd, automation, unit-tests, integration, playwright, jest, nunit |
