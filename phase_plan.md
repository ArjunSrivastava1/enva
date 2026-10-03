# enva v0.3.0 Implementation Plan - Phase 3: Real Security Scanning

## Completed Work (Phase 2: PyPI API Integration)
- [x] Create pkg/api/pypi/client.go
  - Client struct with HTTP client, caching (sync.RWMutex), rate limiting
  - GetVersionInfo(), GetLatestVersion(), GetVersions(), HasVersion(), IsVersionLatest(), GetVersionsSince()
  - Package name normalization (PEP 503 compliant)
  - Version comparison and sorting utilities
- [x] Create pkg/api/pypi/cache.go
  - LRUCache with thread-safe operations
  - TimeBoundedCache with TTL expiration
- [x] Create pkg/api/pypi/options.go
  - Config struct with BaseURL, CacheTTL, RateLimit, MaxCacheSize, Timeout, EnableVersionCache
  - Builder pattern methods for configuration
- [x] Update internal/scanner/dependencies.go
  - Use pypi.NewClient() and client.GetLatestVersion()
  - Add error handling for API calls
  - Export ParseRequirementsFile()
- [x] Update internal/validator/engine.go
  - Use shared PyPI client with context
  - Update scanSecurity() to use API for FixedIn versions
  - Add error handling for security scan API calls
  - Handle API errors gracefully
  - Add rate limiting

## Phase 3: Real Security Scanning (COMPLETED)
- [x] Create pkg/api/osv/vulnerability.go
  - OSV vulnerability data structures (Vulnerability, Reference, Severity, Affected, Package, Range, Event, Distro, Node, Withdrawn, Related)
- [x] Create pkg/api/osv/client.go
  - Client struct with HTTP client, TimeBoundedCache, rate limiting
  - GetVulnerabilities() - fetch vulnerabilities for package+ecosystem
  - GetVulnerability() - fetch specific vulnerability by ID
  - SearchVulnerabilities() - search by keyword
  - GetVulnerabilitiesForPackage() - fetch and convert to internal types
  - ScanAllPackages() - scan all dependencies concurrently
  - Version filtering for installed packages
  - Package name normalization (PEP 503 compliant)
  - Version comparison utilities
- [x] Create pkg/api/osv/mapper.go
  - MapVulnerabilitiesToCVEs() - map OSV IDs to CVEs
  - MapVulnerabilityToSeverity() - map OSV severity to enva severity
  - MapVulnerabilitySeverityScoreTextToEnvaseverity() - map score text to severity
  - MapVulnerabilitySeverityScoreToEnvaseverity() - map score value to severity
  - GetVulnerabilityFixedVersion() - extract fixed version from affected ranges
  - GetVulnerabilityVulnerableVersions() - extract vulnerable version ranges
  - GetVulnerabilityIntroductionVersion() - extract introduction version
- [x] Create pkg/api/osv/cache.go
  - LRUCache for OSV vulnerability data (thread-safe)
  - TimeBoundedCache with TTL expiration for OSV
- [x] Update pkg/api/osv/options.go
  - Config struct with BaseURL, CacheTTL, RateLimit, MaxCacheSize, Timeout, EnableVulnerabilityCache, DefaultEcosystem, MaxConcurrentRequests, EnableCVEMapping, EnableSeverityMapping
  - Builder pattern methods for all configuration options
- [x] Update internal/validator/engine.go
  - Integrate OSV client for real vulnerability scanning
  - Replace hardcoded vulnerable versions with OSV API calls
  - Support concurrent vulnerability scanning
  - Proper version matching against installed packages
  - Map OSV severities to enva severity levels
  - Add CVE ID mapping from OSV references
  - Add helper functions for severity score mapping
- [x] Update internal/types/types.go
  - Add CVEs field to Vulnerability struct for mapped CVE IDs
- [x] Update internal/scanner/security.go
  - Remove hardcoded vulnerable package versions
  - Use OSV client for real-time vulnerability lookup
  - Integrate with OSV API client
  - Support version range filtering

## Phase 4: Enhanced CLI
- [ ] Add --compare PATH1 PATH2
- [ ] Add --export FORMAT (json/html/yaml)
- [ ] Add --format FORMAT (text/json/table)
- [ ] Add --min-score N
- [ ] Add --production-ready flag
- [ ] Add --ignore PKG1,PKG2
- [ ] Update help text

## Phase 5: HTML Report Generation
- [x] Create internal/formatter/html.go
- [x] Generate styled HTML report
  - Include summary dashboard with score cards and status indicators
  - List issues with severity indicators (critical, high, medium, low)
  - Display dependencies with status (uptodate, outdated, vulnerable, missing)
  - Show security vulnerabilities with CVE mappings
  - Include performance analysis (unused packages, large packages, optimization suggestions)
  - Responsive design with mobile-friendly layout
  - Print-friendly styles
  - Professional gradient header with metadata

## Phase 6: Conda & UV Support
- [x] Create internal/detectors/conda.go
  - Check conda-meta/history
  - Parse conda package list
  - Detect conda environments
- [x] Create internal/detectors/uv.go
  - Check .python-version
  - Parse uv.lock
  - Detect UV environments
- [x] Update main.go to support multiple venv types

## Phase 7: Real Performance Analysis
- [ ] Actually measure package sizes from installed files
- [ ] Scan site-packages directory
- [ ] Detect unused packages via import analysis
- [ ] Better optimization suggestions based on real data

## Phase 8: Tests & Documentation
- [ ] Add unit tests for all components
- [ ] Add integration tests with real venvs
- [ ] Add edge case tests
- [ ] Add benchmark tests
- [ ] Update README with comprehensive examples
- [ ] Add CLI help documentation
- [ ] Update LICENSE and metadata

---

## Next Steps
Phase 3 is now complete. Phase 5: HTML Report Generation is now complete. Phase 6: Conda & UV Support is in progress.
All Phase 5 tasks have been completed:
- [x] Create internal/formatter/html.go
- [x] Generate styled HTML report

All Phase 6 tasks have been completed:
- [x] Create internal/detectors/conda.go (completed earlier)
- [x] Create internal/detectors/uv.go (just completed)
- [x] Update main.go to support multiple venv types (just completed)
- [x] Update internal/validator/engine.go to support conda and UV environments (just completed)
- [x] Update internal/types/types.go to add UVEnv type (just completed)
