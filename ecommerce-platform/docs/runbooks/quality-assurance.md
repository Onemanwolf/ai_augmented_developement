# Quality Assurance Runbook

This runbook provides guidance for maintaining code quality, running quality gates, and responding to quality issues in the E-Commerce Platform.

## Table of Contents

1. [Quality Gates Overview](#quality-gates-overview)
2. [Running Quality Checks](#running-quality-checks)
3. [Interpreting Results](#interpreting-results)
4. [Responding to Failures](#responding-to-failures)
5. [Performance Testing](#performance-testing)
6. [Security Scanning](#security-scanning)
7. [Continuous Monitoring](#continuous-monitoring)

---

## Quality Gates Overview

### Thresholds

| Metric | Threshold | Severity |
|--------|-----------|----------|
| Code Coverage | >= 80% | High |
| Lint Warnings | 0 | Medium |
| Security Vulnerabilities | 0 Critical/High | Critical |
| Cyclomatic Complexity | <= 15 | Medium |
| p95 Latency | < 500ms | High |
| Error Rate | < 1% | Critical |

### Quality Gate Jobs

1. **Code Quality** - Linting, complexity analysis
2. **Test Coverage** - Unit tests with coverage reporting
3. **Security Scan** - Vulnerability scanning, secrets detection
4. **Performance Tests** - Load testing with k6
5. **Dependency Review** - License compliance, outdated packages

---

## Running Quality Checks

### Local Development

```bash
# Run all quality gates
./scripts/quality/run-quality-gates.sh

# Run coverage check only
./scripts/quality/check-coverage.sh

# Run security scan only
./scripts/quality/security-scan.sh

# Run load tests (requires services running)
./scripts/quality/run-load-tests.sh smoke
```

### CI/CD Pipeline

Quality gates run automatically on:
- Push to `main` or `develop` branches
- Pull requests to `main` or `develop`
- Daily scheduled runs (midnight UTC)

### Manual Trigger

```bash
# Trigger quality gates workflow manually
gh workflow run quality-gates.yml
```

---

## Interpreting Results

### Coverage Report

```bash
# View coverage summary
go tool cover -func=coverage.out

# Generate HTML report
go tool cover -html=coverage.out -o coverage.html
```

**Understanding Coverage:**
- `statements` - Line coverage
- `branches` - Decision coverage
- Look for uncovered error handling paths

### Lint Results

Common issues and fixes:

| Issue | Solution |
|-------|----------|
| `unused variable` | Remove or use the variable |
| `error not checked` | Add `if err != nil` handling |
| `function too long` | Refactor into smaller functions |
| `cognitive complexity` | Simplify nested logic |

### Security Scan Results

**Critical Issues (Fix Immediately):**
- SQL Injection vulnerabilities
- Command injection
- Hardcoded credentials
- Known CVEs in dependencies

**High Issues (Fix Within 24h):**
- Insecure cryptographic functions
- Missing input validation
- Sensitive data exposure

---

## Responding to Failures

### Coverage Below Threshold

1. **Identify uncovered code:**
   ```bash
   go tool cover -html=coverage.out
   ```

2. **Add missing tests:**
   - Focus on error handling paths
   - Test edge cases
   - Add integration tests if needed

3. **Consider if code is necessary:**
   - Remove dead code
   - Simplify complex functions

### Lint Failures

1. **Auto-fix where possible:**
   ```bash
   golangci-lint run --fix
   ```

2. **Review remaining issues manually**

3. **If false positive, add nolint directive:**
   ```go
   //nolint:errcheck // Intentionally ignoring error
   ```

### Security Vulnerabilities

1. **For dependency vulnerabilities:**
   ```bash
   go get -u <package>@latest
   go mod tidy
   ```

2. **For code vulnerabilities:**
   - Fix the vulnerable code pattern
   - Add input validation
   - Use parameterized queries

3. **If cannot fix immediately:**
   - Create tracking issue
   - Add to technical debt backlog
   - Document mitigation measures

### Performance Failures

1. **Identify bottleneck:**
   - Check database queries
   - Review external service calls
   - Profile with pprof

2. **Common fixes:**
   - Add database indexes
   - Implement caching
   - Optimize algorithms
   - Add pagination

---

## Performance Testing

### Test Types

| Type | Purpose | Duration | VUs |
|------|---------|----------|-----|
| Smoke | Quick validation | 1m | 1 |
| Load | Normal traffic | 15m | 50-100 |
| Stress | Find limits | 25m | 100-300 |
| Spike | Traffic spikes | 10m | 100-500 |

### Running Tests

```bash
# Smoke test (quick validation)
./scripts/quality/run-load-tests.sh smoke

# Full load test
./scripts/quality/run-load-tests.sh load

# Stress test
./scripts/quality/run-load-tests.sh stress

# All tests sequentially
./scripts/quality/run-load-tests.sh all
```

### Analyzing Results

```bash
# View summary
cat reports/load/smoke-results.json | jq '.metrics'

# Check thresholds
cat reports/load/smoke-results.json | jq '.thresholds'
```

### Key Metrics to Monitor

- `http_req_duration` - Request latency
- `http_req_failed` - Error rate
- `iterations` - Throughput
- `vus` - Virtual users handled

---

## Security Scanning

### Automated Scans

1. **govulncheck** - Go vulnerability database
2. **gosec** - Static analysis for security issues
3. **Trivy** - Container and dependency scanning
4. **Dependency Review** - License compliance

### Manual Security Review

For critical changes, perform:

1. **Input Validation Review:**
   - All user inputs sanitized
   - SQL queries parameterized
   - File paths validated

2. **Authentication/Authorization:**
   - Proper token validation
   - Role-based access control
   - Session management

3. **Data Protection:**
   - Sensitive data encrypted
   - Secure communication (TLS)
   - Proper logging (no secrets)

---

## Continuous Monitoring

### Quality Dashboards

- **Codecov**: Code coverage trends
- **GitHub Security**: Vulnerability alerts
- **Grafana**: Performance metrics

### Alerts

Quality alerts are sent to:
- `#ecommerce-platform` Slack channel
- Quality team email distribution

### Weekly Quality Review

Every Monday:
1. Review quality gate trends
2. Address any degradation
3. Update technical debt backlog
4. Plan improvement tasks

### Monthly Security Review

First Monday of each month:
1. Review dependency updates
2. Run full security scan
3. Update security documentation
4. Review access controls

---

## Escalation

### Quality Issues

| Severity | Response Time | Escalation |
|----------|--------------|------------|
| Critical | 1 hour | Tech Lead + Security |
| High | 4 hours | Tech Lead |
| Medium | 24 hours | Team |
| Low | Sprint | Backlog |

### Contacts

- **Quality Lead**: quality@example.com
- **Security Team**: security@example.com
- **On-Call**: oncall@example.com

---

## References

- [Go Testing Guide](https://go.dev/doc/tutorial/add-a-test)
- [golangci-lint](https://golangci-lint.run/)
- [k6 Documentation](https://k6.io/docs/)
- [OWASP Go Security](https://owasp.org/www-project-go-secure-coding-practices-guide/)
