# Security Policy

## Supported Versions

| Version | Supported |
|---------|-----------|
| latest  | Yes       |

## Reporting a Vulnerability

If you discover a security vulnerability in tfwatch, please report it responsibly.

**Do not open a public GitHub issue for security vulnerabilities.**

Instead, please email **security@cloudpulse-hq.com** with:

- A description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix (if any)

You should receive an acknowledgment within **48 hours**. We will work with you to understand the issue and coordinate a fix and disclosure timeline.

## Scope

tfwatch is a CLI tool that reads local Terraform files and publishes metrics. It does not accept network input or run a server. Security concerns most likely involve:

- Unintended information disclosure from parsed Terraform configs
- Dependency supply chain issues (Go modules)
- OTEL endpoint credential handling
