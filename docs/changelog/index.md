---
title: Detailed Changelog
sidebar_position: 0
---

# Detailed Changelog

Welcome to the zrok detailed changelog. This section provides comprehensive release notes for each zrok version, including:

- Feature descriptions and usage examples
- Breaking changes and migration guidance
- Technical implementation details
- Bug fixes and improvements
- Links to related documentation

## Quick vs. Detailed Changelogs

- **Quick Summary**: For a concise list of changes across all versions, see the [CHANGELOG.md](https://github.com/openziti/zrok/blob/main/CHANGELOG.md) in the repository root.
- **Detailed Release Notes**: The pages in this section provide in-depth information about major changes, with examples and context for both end users and self-hosters.

:::note
Detailed changelogs began with v2.0.0. For earlier versions, please refer to the [CHANGELOG.md](https://github.com/openziti/zrok/blob/main/CHANGELOG.md).
:::

## Available Detailed Changelogs

### Version 2.x

- [**v2.0.0**](./v2.0.0.md) - major paradigm shift: namespaces and names replace reserved shares, new dynamic proxy system, agent improvements, and extensive API changes

---

## Understanding zrok Versioning

zrok follows [Semantic Versioning](https://semver.org/):

- **Major versions** (e.g., 2.0.0) introduce breaking changes that may require migration steps
- **Minor versions** (e.g., 2.1.0) add new features while maintaining backward compatibility
- **Patch versions** (e.g., 2.0.1) contain bug fixes and small improvements

### Upgrading Guidance

When upgrading zrok:

1. **Review the detailed changelog** for your target version to understand what's changing
2. **Check for breaking changes** that might affect your setup
3. **Follow migration guides** if moving between major versions
4. **Test in a non-production environment** when possible
5. **Backup your data** before upgrading self-hosted instances

For self-hosters, pay special attention to:
- Database migrations (handled automatically by default, but can be managed with `zrok admin migrate`)
- Configuration changes (some versions require config updates)
- Infrastructure requirements (new components like RabbitMQ in v2.0.0)

---

## Contributing

Found an error or want to improve the documentation? Contributions are welcome!

- Report issues: [GitHub Issues](https://github.com/openziti/zrok/issues)
- Suggest improvements: [OpenZiti Discourse](https://openziti.discourse.group/)
- Submit changes: [GitHub Pull Requests](https://github.com/openziti/zrok/pulls)
