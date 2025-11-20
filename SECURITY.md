# Security Policy

## Supported Versions

We release patches for security vulnerabilities in the following versions:

| Version | Supported          |
| ------- | ------------------ |
| 1.0.x   | :white_check_mark: |

## Reporting a Vulnerability

We take the security of StreamLens seriously. If you have discovered a security vulnerability, please follow these guidelines:

### Please Do

- **Report privately**: Report security vulnerabilities by opening a private security advisory on GitHub or by emailing the maintainers directly (create a GitHub issue with `@maintainer` mention if email is not available)
- **Provide details**: Include as much information as possible:
  - Type of vulnerability (e.g., SQL injection, XSS, authentication bypass)
  - Affected component and version
  - Steps to reproduce the issue
  - Potential impact
  - Any suggested fixes (optional)
- **Allow time**: Give us reasonable time to address the issue before public disclosure (we aim for 90 days)
- **Act in good faith**: Make a good faith effort to avoid privacy violations, data destruction, and service interruption

### Please Don't

- **Don't disclose publicly**: Don't open public GitHub issues for security vulnerabilities
- **Don't test on production**: Don't test vulnerabilities on production deployments you don't own
- **Don't exploit**: Don't exploit the vulnerability beyond what's necessary to demonstrate it exists

## Response Process

1. **Acknowledgment**: We'll acknowledge receipt of your vulnerability report within 48 hours
2. **Assessment**: We'll investigate and validate the reported vulnerability
3. **Fix**: We'll develop and test a fix
4. **Disclosure**: We'll coordinate with you on a responsible disclosure timeline
5. **Release**: We'll release a security patch and credit you (unless you prefer to remain anonymous)

## Security Best Practices

When deploying StreamLens in production:

### Network Security
- **Use HTTPS**: Always use HTTPS for the Ingestion API and Metrics API
- **Network isolation**: Run services in a private network
- **Firewall rules**: Restrict access to ports (only expose necessary endpoints)
- **API Gateway**: Consider placing APIs behind an API gateway with rate limiting

### Authentication & Authorization
- **Add authentication**: The default deployment has no authentication - add API keys, OAuth, or JWT
- **Tenant isolation**: Ensure proper tenant isolation in queries
- **Database access**: Restrict database access to only the application user
- **Secrets management**: Use a secrets manager (not environment variables) for production

### Infrastructure
- **Update regularly**: Keep dependencies and base images up to date
- **Least privilege**: Run containers with minimal privileges (non-root user)
- **Resource limits**: Set CPU and memory limits to prevent DoS
- **Logging**: Enable comprehensive logging and monitoring
- **Backup**: Regularly backup PostgreSQL data

### Data Privacy
- **PII handling**: Be careful with user data in metadata fields
- **Encryption at rest**: Enable encryption for PostgreSQL volumes
- **Encryption in transit**: Use TLS for Kafka and Postgres connections
- **Data retention**: Implement a data retention policy

### Kafka/Redpanda Security
- **Enable SASL/SCRAM**: Use authentication for Kafka connections
- **Enable TLS**: Encrypt Kafka traffic
- **Topic ACLs**: Restrict topic access per service
- **Consumer groups**: Use unique consumer group IDs per environment

### Container Security
- **Scan images**: Regularly scan Docker images for vulnerabilities
- **Minimal base images**: Use minimal base images (alpine, distroless)
- **No secrets in images**: Never bake secrets into container images
- **Read-only filesystem**: Where possible, use read-only container filesystems

## Known Security Considerations

### Current Limitations
- **No built-in authentication**: The default setup has no API authentication
- **No rate limiting**: The ingestion API has no built-in rate limiting
- **In-memory state**: The processor keeps join state in memory (potential data loss on crash)
- **No encryption by default**: Database and Kafka use unencrypted connections

These are known trade-offs for development simplicity. For production deployments, address these using the best practices above.

## Security Updates

Security updates will be:
- Released as patch versions (e.g., 1.0.1)
- Announced in the CHANGELOG with a `[SECURITY]` tag
- Documented in GitHub Security Advisories
- Highlighted in the project README

## Questions?

If you have questions about security that don't involve reporting a vulnerability, please open a GitHub issue with the `security` label.

Thank you for helping keep StreamLens and its users safe!
