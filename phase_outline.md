enva v0.2.0 Upgrade Roadmap
============================

Phase 2: PyPI API Integration (COMPLETED)
-------------------------------------------
- Create pkg/api/pypi/client.go ✓
  - Client struct with HTTP client, caching (sync.RWMutex), rate limiting
  - GetVersionInfo(), GetLatestVersion(), GetVersions(), HasVersion(), IsVersionLatest(), GetVersionsSince()
  - Package name normalization (PEP 503 compliant)
  - Version comparison and sorting utilities
- Create pkg/api/pypi/cache.go ✓
  - LRUCache with thread-safe operations
  - TimeBoundedCache with TTL expiration
- Create pkg/api/pypi/options.go ✓
  - Config struct with BaseURL, CacheTTL, RateLimit, MaxCacheSize, Timeout, EnableVersionCache
  - Builder pattern methods for configuration
- Update internal/scanner/dependencies.go ✓
  - Use pypi.NewClient() and client.GetLatestVersion()
  - Add error handling for API calls
  - Export ParseRequirementsFile()
- Update internal/validator/engine.go ✓
  - Use shared PyPI client with context
  - Update scanSecurity() to use API for FixedIn versions
  - Add error handling for security scan API calls
- Handle API errors gracefully ✓
- Add rate limiting ✓

Phase 3: Real Security Scanning
-------------------------------
- Integrate OSV.dev API or PyPI vulnerability API
- Fetch vulnerabilities for all installed packages
- Map OSV IDs to CVEs
- Update vulnerability database with real data
- Support multiple severity levels

Phase 4: Enhanced CLI
---------------------
- Add --compare PATH1 PATH2
- Add --export FORMAT (json/html/yaml)
- Add --format FORMAT (text/json/table)
- Add --min-score N
- Add --production-ready flag
- Add --ignore PKG1,PKG2
- Update help text

Phase 5: HTML Report Generation
-------------------------------
- Create internal/formatter/html.go
- Generate styled HTML report
- Include summary dashboard
- List issues with severity indicators
- Show comparison charts (if --compare used)

Phase 6: Conda & UV Support
---------------------------
- Create internal/detectors/conda.go
  - Check conda-meta/history
  - Parse conda package list
  - Detect conda environments
- Create internal/detectors/uv.go
  - Check .python-version
  - Parse uv.lock
  - Detect UV environments
- Update main.go to support multiple venv types

Phase 7: Real Performance Analysis
----------------------------------
- Actually measure package sizes from installed files
- Scan site-packages directory
- Detect unused packages via import analysis
- Better optimization suggestions based on real data

Phase 8: Tests & Documentation
------------------------------
- Add unit tests for all components
- Add integration tests with real venvs
- Add edge case tests
- Add benchmark tests
- Update README with comprehensive examples
- Add CLI help documentation
- Update LICENSE and metadata