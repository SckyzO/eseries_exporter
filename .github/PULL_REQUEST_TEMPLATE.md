## Description

Please include a summary of the changes and the related issue. Include relevant motivation and context.

Fixes # (issue number)

## Type of Change

What types of changes does your code introduce to eseries_exporter? Put an `x` in the boxes that apply.

- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update
- [ ] Performance improvement
- [ ] Refactoring
- [ ] Tests
- [ ] CI/CD changes

## Checklist

Put an `x` in the boxes that apply. You can also fill these out after creating the PR. If you're unsure about any of them, don't hesitate to ask. We're here to help!

- [ ] My code follows the style guidelines of this project
- [ ] I have performed a self-review of my code
- [ ] I have commented my code, particularly in hard-to-understand areas
- [ ] I have made corresponding changes to the documentation
- [ ] My changes generate no new warnings
- [ ] I have added tests that prove my fix is effective or that my feature works
- [ ] New and existing unit tests pass locally with my changes
- [ ] Any dependent changes have been merged and published

## Testing

How has this been tested? Please describe the tests that you ran to verify your changes.

### Manual Testing

- [ ] I tested this change manually
- [ ] Verified the changes work as expected
- [ ] Tested with different configurations
- [ ] Tested with different data sets

### Automated Testing

- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Code coverage remains acceptable
- [ ] Performance benchmarks are met

## Breaking Changes

Does this PR introduce any breaking changes? If yes, please describe:

- What functionality will break
- How users will be affected
- Migration guide for users
- Any workarounds

## Performance Impact

Describe any performance impact and how it was tested.

- CPU usage impact: <!-- High/Medium/Low/None -->
- Memory usage impact: <!-- High/Medium/Low/None -->
- Network usage impact: <!-- High/Medium/Low/None -->
- Load testing results: <!-- Link to results -->

## Security Impact

Describe any security-related changes.

- [ ] No security impact
- [ ] Security improvements
- [ ] TLS/SSL changes
- [ ] Authentication changes
- [ ] Authorization changes

## Documentation Changes

What documentation needs to be updated?

- [ ] README.md
- [ ] CHANGELOG.md
- [ ] Code comments
- [ ] API documentation
- [ ] Examples
- [ ] Configuration documentation

## Configuration Examples

If this change affects configuration, provide example configurations:

```yaml
# Before
modules:
  default:
    user: monitor
    password: secret
    proxy_url: http://localhost:8080

# After
modules:
  default:
    user: monitor
    password: secret
    proxy_url: http://localhost:8080
    # New option added
    new_option: value
```

## Metrics Changes

If this change affects metrics, list the changes:

- New metrics added:
- Metrics modified:
- Metrics removed:

## Related Issues

Link any related issues:

- Related Issue: #ISSUE_NUMBER
- Related PR: #PR_NUMBER
- Related Discussion: #DISCUSSION_NUMBER

## Screenshots

If applicable, add screenshots to help explain your changes.

## Additional Notes

Add any other notes or context about the pull request here.

## Reviewers

Tag specific reviewers you want to review this PR:

- [ ] @reviewer1
- [ ] @reviewer2

## Labels

The following labels help categorize this PR:

- [ ] bug
- [ ] enhancement
- [ ] documentation
- [ ] performance
- [ ] refactoring
- [ ] security
- [ ] testing
- [ ] needs-review
- [ ] ready-for-review