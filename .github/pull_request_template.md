## Summary

<!-- Provide a concise description of the change and its motivation -->

## Type of Change

- [ ] Bug fix (non-breaking change that fixes an issue)
- [ ] New feature (non-breaking change that adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Refactor (no functional changes)
- [ ] Documentation update
- [ ] Proto/API contract change

## Related Issues

<!-- Link to relevant issues: Fixes #123, Relates to #456 -->

## Checklist

- [ ] My code follows the project's Clean Architecture boundaries
- [ ] I have added tests that prove my fix/feature works
- [ ] All new and existing tests pass (`make test`)
- [ ] I have run `make lint` and resolved all warnings
- [ ] Proto changes have been reviewed by the lead architect
- [ ] No business logic exists in handler layer
- [ ] All external calls have OTel spans
- [ ] Errors are wrapped with context (`fmt.Errorf("pkg.Func: %w", err)`)
- [ ] Context is propagated through all I/O functions
- [ ] No cross-package imports between internal systems

## Architecture Impact

<!-- If this changes system boundaries, data flow, or interfaces, describe the impact -->

## Screenshots / Logs

<!-- If applicable, add screenshots or relevant log output -->
