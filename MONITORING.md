
# Monitoring & Alerting Guide

This guide covers monitoring setup, alerting rules, and dashboards for the Deployment Tail application with authentication.

## Table of Contents

1. [Overview](#overview)
2. [Key Metrics](#key-metrics)
3. [Prometheus Configuration](#prometheus-configuration)
4. [Alert Rules](#alert-rules)
5. [Grafana Dashboards](#grafana-dashboards)
6. [Log Monitoring](#log-monitoring)
7. [Security Monitoring](#security-monitoring)
8. [Incident Response](#incident-response)

## Overview

### Monitoring Objectives

1. **Availability**: Ensure the service is up and responding
2. **Performance**: Track response times and throughput
3. **Security**: Detect authentication failures and suspicious activity
4. **User Experience**: Monitor user registrations and active sessions
5. **Capacity**: Track database and token growth

### Monitoring Stack

- **Metrics**: Prometheus
- **Alerts**: Alertmanager
- **Dashboards**: Grafana
- **Logs**: ELK Stack (Elasticsearch, Logstash, Kibana) or Loki
- **APM**: Optional (Datadog, New Relic, etc.)

## Key Metrics

### Authentication Metrics

```prometheus
# Total authentication attempts
auth_attempts_total{result="success|failure",reason=""}

# Authentication failures by reason
auth_failures_total{reason="invalid_token|expired_token|revoked_token|missing_token|invalid_signature"}

# Token operations
token_issued_total
token_refreshed_total
token_revoked_total

# OAuth operations
oauth_login_attempts_total{result="success|failure"}
oauth_callback_duration_seconds

# JWT validation
jwt_validation_duration_seconds
jwt_validation_errors_total{reason=""}
```

### User Metrics

```prometheus
# User registrations
user_registrations_total
user_registrations_rate (rate of new users)

# Active sessions
active_sessions_gauge
active_sessions_by_role{role="viewer|deployer|admin"}

# Last login tracking
user_last_login_seconds_ago

# Role distribution
users_by_role_gauge{role="viewer|deployer|admin"}
```

### API Performance

```prometheus
# HTTP requests
http_requests_total{method,endpoint,status}
http_request_duration_seconds{method,endpoint}

# Schedule operations
schedule_operations_total{operation="create|update|delete|approve|deny"}
schedule_operation_duration_seconds{operation}
```

### Database Metrics

```prometheus
# Database queries
db_queries_total{operation="select|insert|update|delete"}
db_query_duration_seconds{operation}

# Token revocation table
revoked_tokens_count
revoked_tokens_cleanup_duration_seconds

# Users table
users_table_size_bytes
users_count
```

### Security Metrics

```prometheus
# Brute force detection
auth_failures_per_user_total{email}
auth_failures_per_ip_total{ip}

# Suspicious activity
role_escalation_attempts_total
unauthorized_access_attempts_total{endpoint}

# Token abuse
token_reuse_after_revocation_total
expired_token_usage_attempts_total
```

## Prometheus Configuration

### Scrape Configuration

Add this to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'deployment-tail'
    scrape_interval: 15s
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: /metrics

    # Optional: Add labels
    relabel_configs:
      - source_labels: [__address__]
        target_label: instance
        regex: '([^:]+):.*'
        replacement: '${1}'
```

### Recording Rules

Add this to `prometheus_rules.yml`:

```yaml
groups:
  - name: authentication
    interval: 30s
    rules:
      # Authentication failure rate (per minute)
      - record: auth:failures:rate1m
        expr: rate(auth_failures_total[1m])

      # Authentication success rate
      - record: auth:success_rate
        expr: |
          sum(rate(auth_attempts_total{result="success"}[5m]))
          /
          sum(rate(auth_attempts_total[5m]))

      # Active sessions growth rate
      - record: users:active_sessions:growth_rate5m
        expr: deriv(active_sessions_gauge[5m])

  - name: performance
    interval: 30s
    rules:
      # 95th percentile response time
      - record: http:request_duration:p95
        expr: histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))

      # Error rate
      - record: http:error_rate
        expr: |
          sum(rate(http_requests_total{status=~"5.."}[5m]))
          /
          sum(rate(http_requests_total[5m]))
```

## Alert Rules

### Critical Alerts

```yaml
groups:
  - name: critical
    rules:
      # Service down
      - alert: ServiceDown
        expr: up{job="deployment-tail"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Deployment Tail service is down"
          description: "Service has been down for more than 1 minute"

      # No active admins
      - alert: NoAdminUsers
        expr: users_by_role_gauge{role="admin"} == 0
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "No admin users available"
          description: "System has no admin users. Admin actions are blocked."

      # High error rate
      - alert: HighErrorRate
        expr: http:error_rate > 0.05
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "High API error rate ({{ $value | humanizePercentage }})"
          description: "More than 5% of requests are failing"

      # Database connection failure
      - alert: DatabaseConnectionFailure
        expr: up{job="mysql"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Cannot connect to database"
          description: "Database connection has been down for 1 minute"
```

### Authentication Alerts

```yaml
groups:
  - name: authentication
    rules:
      # High authentication failure rate
      - alert: HighAuthFailureRate
        expr: auth:failures:rate1m > 10
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High authentication failure rate"
          description: "More than 10 auth failures per minute for 5 minutes"

      # Potential brute force attack
      - alert: PossibleBruteForceAttack
        expr: |
          sum by (email) (
            increase(auth_failures_per_user_total[5m])
          ) > 20
        labels:
          severity: critical
        annotations:
          summary: "Possible brute force attack on user {{ $labels.email }}"
          description: "More than 20 failed login attempts in 5 minutes"

      # OAuth provider issues
      - alert: OAuthFailureRate
        expr: |
          sum(rate(oauth_login_attempts_total{result="failure"}[5m]))
          /
          sum(rate(oauth_login_attempts_total[5m])) > 0.5
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High OAuth failure rate"
          description: "More than 50% of OAuth attempts failing"

      # Token revocation table growing too large
      - alert: RevokedTokensTableTooLarge
        expr: revoked_tokens_count > 100000
        labels:
          severity: warning
        annotations:
          summary: "Revoked tokens table is too large"
          description: "Table has {{ $value }} entries. Consider cleanup."

      # JWT validation slow
      - alert: JWTValidationSlow
        expr: |
          histogram_quantile(0.95,
            rate(jwt_validation_duration_seconds_bucket[5m])
          ) > 0.1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "JWT validation is slow (p95: {{ $value }}s)"
          description: "95th percentile validation time exceeds 100ms"
```

### Security Alerts

```yaml
groups:
  - name: security
    rules:
      # Revoked token reuse attempts
      - alert: RevokedTokenReuseAttempts
        expr: increase(token_reuse_after_revocation_total[5m]) > 5
        labels:
          severity: warning
        annotations:
          summary: "Attempts to reuse revoked tokens detected"
          description: "{{ $value }} attempts to use revoked tokens in 5 minutes"

      # Role escalation attempts
      - alert: RoleEscalationAttempts
        expr: increase(role_escalation_attempts_total[5m]) > 0
        labels:
          severity: critical
        annotations:
          summary: "Role escalation attempts detected"
          description: "{{ $value }} unauthorized role change attempts"

      # Unusual token refresh rate
      - alert: UnusualTokenRefreshRate
        expr: |
          sum(rate(token_refreshed_total[5m])) >
          3 * avg_over_time(sum(rate(token_refreshed_total[5m]))[1h:5m])
        labels:
          severity: warning
        annotations:
          summary: "Unusual token refresh rate"
          description: "Token refresh rate is 3x higher than normal"
```

### Capacity Alerts

```yaml
groups:
  - name: capacity
    rules:
      # High active sessions
      - alert: HighActiveSessions
        expr: active_sessions_gauge > 1000
        labels:
          severity: warning
        annotations:
          summary: "High number of active sessions"
          description: "{{ $value }} active sessions (normal: < 1000)"

      # Rapid user growth
      - alert: RapidUserGrowth
        expr: users:active_sessions:growth_rate5m > 10
        labels:
          severity: info
        annotations:
          summary: "Rapid user growth detected"
          description: "Adding {{ $value }} users per minute"

      # Database table size
      - alert: UsersTableLarge
        expr: users_table_size_bytes > 1073741824  # 1GB
        labels:
          severity: info
        annotations:
          summary: "Users table is large"
          description: "Users table size: {{ $value | humanizeBytes }}"
```

## Grafana Dashboards

### Dashboard 1: Authentication Overview

**Purpose**: Monitor authentication health and user activity

**Panels**:

1. **Success Rate** (Gauge)
   ```promql
   auth:success_rate * 100
   ```

2. **Failed Logins (Last Hour)** (Single Stat)
   ```promql
   sum(increase(auth_failures_total[1h]))
   ```

3. **Authentication Attempts** (Graph)
   ```promql
   sum(rate(auth_attempts_total[5m])) by (result)
   ```

4. **Failure Reasons** (Pie Chart)
   ```promql
   sum(increase(auth_failures_total[1h])) by (reason)
   ```

5. **Active Sessions** (Graph)
   ```promql
   active_sessions_gauge
   active_sessions_by_role{role=~"viewer|deployer|admin"}
   ```

6. **New User Registrations** (Graph)
   ```promql
   sum(increase(user_registrations_total[1h]))
   ```

7. **Users by Role** (Bar Gauge)
   ```promql
   users_by_role_gauge
   ```

8. **OAuth Success Rate** (Gauge)
   ```promql
   sum(rate(oauth_login_attempts_total{result="success"}[5m]))
   /
   sum(rate(oauth_login_attempts_total[5m])) * 100
   ```

### Dashboard 2: Security Monitoring

**Purpose**: Detect and investigate security incidents

**Panels**:

1. **Brute Force Attempts** (Table)
   ```promql
   topk(10, sum by (email) (
     increase(auth_failures_per_user_total[1h])
   ))
   ```

2. **Revoked Token Reuse** (Single Stat)
   ```promql
   sum(increase(token_reuse_after_revocation_total[24h]))
   ```

3. **Role Escalation Attempts** (Alert List)
   ```promql
   increase(role_escalation_attempts_total[24h])
   ```

4. **Unauthorized Access Attempts** (Heatmap)
   ```promql
   sum(increase(unauthorized_access_attempts_total[1h])) by (endpoint)
   ```

5. **Failed Auth by IP** (Table)
   ```promql
   topk(10, sum by (ip) (
     increase(auth_failures_per_ip_total[1h])
   ))
   ```

6. **Expired Token Usage** (Graph)
   ```promql
   sum(rate(expired_token_usage_attempts_total[5m]))
   ```

### Dashboard 3: Performance & Capacity

**Purpose**: Track system performance and capacity

**Panels**:

1. **Response Time (p95)** (Graph)
   ```promql
   histogram_quantile(0.95,
     sum(rate(http_request_duration_seconds_bucket[5m])) by (le, endpoint)
   )
   ```

2. **JWT Validation Time** (Graph)
   ```promql
   histogram_quantile(0.95,
     rate(jwt_validation_duration_seconds_bucket[5m])
   )
   ```

3. **Database Query Duration** (Graph)
   ```promql
   histogram_quantile(0.95,
     rate(db_query_duration_seconds_bucket[5m])
   ) by (operation)
   ```

4. **Revoked Tokens Count** (Graph)
   ```promql
   revoked_tokens_count
   ```

5. **Token Operations Rate** (Graph)
   ```promql
   sum(rate(token_issued_total[5m]))
   sum(rate(token_refreshed_total[5m]))
   sum(rate(token_revoked_total[5m]))
   ```

6. **Error Rate** (Graph)
   ```promql
   http:error_rate * 100
   ```

### Importing Dashboards

```bash
# Export dashboard JSON
curl -H "Authorization: Bearer $GRAFANA_TOKEN" \
  http://grafana:3000/api/dashboards/db/authentication-overview

# Import dashboard
curl -X POST -H "Content-Type: application/json" \
  -H "Authorization: Bearer $GRAFANA_TOKEN" \
  -d @dashboard.json \
  http://grafana:3000/api/dashboards/db
```

## Log Monitoring

### Important Log Patterns

#### Authentication Events

```
# Successful login
level=info msg="User authenticated successfully" user_id=123 email=user@example.com google_id=456

# Failed login
level=warn msg="Authentication failed" reason="invalid_token" ip=192.168.1.100

# Token revoked
level=info msg="Token revoked" user_id=123 reason="logout"

# Admin created
level=info msg="Admin user created" user_id=123 email=admin@example.com
```

#### Security Events

```
# Brute force detected
level=warn msg="Possible brute force attack" email=victim@example.com attempt_count=25 ip=10.0.0.50

# Role escalation attempt
level=error msg="Unauthorized role change attempt" user_id=123 requested_role="admin" current_role="viewer"

# Revoked token reuse
level=warn msg="Revoked token used" token_hash=abc123 user_id=456
```

### ELK Stack Queries

**Find failed authentications:**
```
message:"Authentication failed" AND @timestamp:[now-1h TO now]
```

**Detect brute force:**
```
message:"Authentication failed" AND ip.keyword:*
| stats count by ip | where count > 10
```

**Track admin actions:**
```
message:"Admin action" AND user.role:"admin"
```

### Loki Queries

```logql
# Failed auth in last hour
{job="deployment-tail"} |= "Authentication failed" | logfmt

# Brute force detection
sum by (ip) (
  count_over_time({job="deployment-tail"} |= "Authentication failed" | logfmt [5m])
) > 10

# Admin actions
{job="deployment-tail"} |= "Admin action" | json | role="admin"
```

## Security Monitoring

### Automated Security Checks

Create a cron job for daily security audits:

```bash
#!/bin/bash
# /etc/cron.daily/deployment-tail-security-audit

# Check for inactive admin accounts
mysql deployment_schedules << EOF
SELECT email, last_login_at
FROM users
WHERE role = 'admin'
AND (last_login_at IS NULL OR last_login_at < DATE_SUB(NOW(), INTERVAL 90 DAY));
EOF

# Check for users with multiple failed logins
# (Query Prometheus or logs)

# Check revoked tokens table size
mysql deployment_schedules << EOF
SELECT COUNT(*) as revoked_count,
       COUNT(*) * 100.0 / (SELECT COUNT(*) FROM users) as ratio
FROM revoked_tokens;
EOF

# Email report to security team
mail -s "Daily Security Audit" security@company.com < audit_report.txt
```

### Security Dashboard Checklist

- [ ] Failed login attempts (last 24h)
- [ ] Revoked token reuse attempts
- [ ] Role escalation attempts
- [ ] Admin users without recent login
- [ ] Users with excessive permissions
- [ ] Unusual token refresh patterns
- [ ] Geographic anomalies (if tracking IP)

## Incident Response

### Authentication Incident Playbook

**Scenario 1: High Authentication Failure Rate**

1. Check alert: `HighAuthFailureRate`
2. Identify source:
   ```bash
   # Check Prometheus
   sum by (email) (increase(auth_failures_per_user_total[5m]))

   # Check logs
   grep "Authentication failed" /var/log/deployment-tail/app.log | tail -100
   ```
3. Determine cause:
   - Single user? Password reset needed
   - Multiple users? OAuth provider issue
   - Single IP? Possible attack
4. Take action:
   - If attack: Block IP in firewall
   - If OAuth issue: Check Google OAuth status
   - If user issue: Contact user

**Scenario 2: Admin Lockout**

1. Check alert: `NoAdminUsers`
2. Verify:
   ```sql
   SELECT * FROM users WHERE role = 'admin';
   ```
3. Create emergency admin:
   ```bash
   export ADMIN_EMAIL=emergency@company.com
   go run scripts/seed_admin_user.go
   ```
4. Document incident
5. Review why admins were deleted/demoted

**Scenario 3: OAuth Provider Outage**

1. Check OAuth success rate
2. Verify Google OAuth status page
3. If prolonged outage:
   ```bash
   # Emergency: Disable authentication
   ./scripts/emergency_auth_disable.sh
   ```
4. Monitor Google status
5. Re-enable authentication when resolved

## Integration with Incident Management

### PagerDuty Integration

```yaml
# alertmanager.yml
receivers:
  - name: 'pagerduty'
    pagerduty_configs:
      - service_key: '<your-pagerduty-key>'
        description: '{{ .GroupLabels.alertname }}: {{ .CommonAnnotations.summary }}'

route:
  group_by: ['alertname']
  receiver: 'pagerduty'
  routes:
    - match:
        severity: critical
      continue: true
```

### Slack Integration

```yaml
receivers:
  - name: 'slack'
    slack_configs:
      - api_url: '<slack-webhook-url>'
        channel: '#alerts'
        title: '{{ .GroupLabels.alertname }}'
        text: '{{ .CommonAnnotations.description }}'
```

## Maintenance

### Weekly Tasks

- Review authentication metrics trends
- Check for inactive users
- Review security incidents
- Update alert thresholds if needed

### Monthly Tasks

- Audit admin users
- Review and cleanup revoked tokens table
- Analyze authentication patterns
- Update monitoring dashboards
- Test alert notifications

### Quarterly Tasks

- Review all alert rules
- Update monitoring strategy
- Security audit
- Capacity planning review

## Resources

- Prometheus documentation: https://prometheus.io/docs/
- Grafana dashboards: https://grafana.com/grafana/dashboards/
- Alert examples: https://awesome-prometheus-alerts.grep.to/
- Security best practices: https://cheatsheetseries.owasp.org/
