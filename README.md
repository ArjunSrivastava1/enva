# **📄 ENVA README.md (Same Professional Style)**

```markdown
<h1>
  <br>
  <img src="https://raw.githubusercontent.com/ArjunSrivastava1/enva/main/assets/icon.svg" alt="enva" width="100">
  <br>
</h1>

<h4>A Python environment validator with clean, actionable output</h4>

<p>
  <a href="https://golang.org"><img src="https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go&logoColor=white" alt="Go Version"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-GPL%20v2-blue.svg" alt="License"></a>
  <a href="https://github.com/ArjunSrivastava1/enva/releases"><img src="https://img.shields.io/github/v/release/ArjunSrivastava1/enva" alt="Release"></a>
  <a href="https://python.org"><img src="https://img.shields.io/badge/Python-3.7+-3776AB?logo=python&logoColor=white" alt="Python"></a>
</p>

<p>
  <a href="#-about">About</a> •
  <a href="#-features">Features</a> •
  <a href="#-quick-start">Quick Start</a> •
  <a href="#-usage">Usage</a> •
  <a href="#-configuration">Configuration</a> •
  <a href="#-contributing">Contributing</a>
</p>

<p>
  <img src="https://raw.githubusercontent.com/ArjunSrivastava1/enva/main/assets/demo.png" alt="Demo" width="700">
</p>

## About

enva is a command-line tool that inspects Python virtual environments with precision. It validates Python and pip versions, checks installed packages for security vulnerabilities, identifies outdated dependencies, and provides clean, actionable feedback in a Chinese-style output format.

Stop the "works on my machine" problem before it starts.

## ✨ Features

| Category | Features |
|----------|----------|
| **🔍 Inspection** | Python version • Pip version • Package inventory • Venv validation |
| **🛡️ Security** | Vulnerability scanning • Outdated detection • Performance hints • Consistency checks |
| **🎨 Output** | Chinese-style formatting • JSON for CI/CD • Actionable suggestions • Minimal design |

## 🚀 Quick Start

### 📦 Installation
```bash
# One-liner install
go install github.com/ArjunSrivastava1/enva@latest
```

## 🎯 Basic Usage
```bash
# Validate current environment (auto-detects .venv/)
enva

# Specify virtual environment path
enva --venv ./venv

# JSON output for automation/CI
enva --json --venv ./venv

# Quick check (minimal output)
enva --quiet
```

## 🛠️ Usage

### Python Development Workflow
```bash
# Morning environment check
enva

# Pre-commit validation
enva --min-score 80

# Compare development vs production
enva --compare dev/.venv prod/.venv
```

### CI/CD Integration
```bash
# Pipeline validation (exit code 1 if score < 80)
enva --json --venv $VENV_PATH | jq '.score >= 80'

# Production readiness check
enva --production-ready --venv /opt/venv

# Security audit with report
enva --security-scan --export report.json
```

### Team Environment Standards
```yaml
# .enva-team.yaml
required_packages:
  - django>=4.0
  - requests>=2.28
  
security_rules:
  max_critical: 0
  max_medium: 2
  
performance:
  max_package_size_mb: 100
```

## 📊 Example Output

```bash
$ enva --venv .venv

[✅ STATUS] Environment Validation
───────────────────────────────────────────
Time: 14:30:00 | Duration: 3.2s

[✅ VENV] Virtual Environment
  ℹ️ Path: .venv
  ✅ Python: 3.9.6
  ✅ Pip: 25.3
  - Status: activated
  ✅ Integrity: valid

[📦 PACKAGES] Dependencies (12)
  - django==4.2.27
  - pandas==2.3.3
  ⚠️ requests==2.28.2 → latest: 2.31.0
  - numpy==2.0.2

[🛡️ SECURITY] Scan Results
  ⚠️ 1 medium vulnerability
  [ISSUE] CVE-2023-XXXXX in requests<2.29.0
  [FIX] Upgrade to requests>=2.31.0

[📈 SUMMARY] Final Assessment
  Score: 85/100
  Status: WARNING
  Time: 3.2 seconds
  Issues: 1 security, 1 outdated

[⚠️ REVIEW] Environment needs attention
```

## ⚙️ Configuration

### Output Formats
```bash
# Human-readable (default)
enva --venv .venv

# JSON for automation
enva --json --venv .venv

# Minimal output
enva --quiet --venv .venv

# Export to file
enva --export report.html --venv .venv
```

### Validation Rules
```bash
# Minimum score requirement
enva --min-score 80 --venv .venv

# Security-only scan
enva --security-only --venv .venv

# Performance analysis
enva --performance --venv .venv
```

## 🤝 Contributing

1. Fork & clone the repository
2. Create feature branch (`feat/` or `fix/`)
3. Commit with Conventional Commits
4. Push & open Pull Request

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines.

## 📄 License

GPL v2.0 - See [LICENSE](LICENSE)

---

<p align="center">
  Built with ❤️ by <a href="https://github.com/ArjunSrivastava1">Arjun Srivastava</a>
</p>

<p align="center">
  <a href="https://github.com/ArjunSrivastava1/enva/issues">Report Bug</a> • 
  <a href="https://github.com/ArjunSrivastava1/enva/issues">Request Feature</a> •
  <a href="https://github.com/ArjunSrivastava1/commit-linter">commit-linter</a> •
  <a href="https://github.com/ArjunSrivastava1/port-scanner">port-scanner</a>
</p>

<p align="center">
  <em>Stop guessing. Start validating.</em>
</p>