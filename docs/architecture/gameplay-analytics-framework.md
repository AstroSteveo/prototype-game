# Gameplay Analytics Framework

## Overview
- Define a repeatable analytics pipeline that captures gameplay behavior, player flows, and systemic health without degrading the live simulation.
- Establish a shared vocabulary so gameplay, economy, and live-ops stakeholders interpret events the same way.
- Deliver dashboards and alerts that pair real-time operational telemetry with longer-term product insights.
- Align with existing WebSocket transport (`/ws`) and simulation tick loop so instrumentation evolves with the prototype roadmap.

## Goals
### Product Goals
- Measure onboarding funnel from first session through first combat win within seven days of activation.
- Quantify skill, economy, and encounter loops to validate prototype hypotheses before the next roadmap checkpoint (2025-10-31).
- Empower designers with self-serve cohort slices (class, region, playstyle) for sprint reviews.

### Telemetry Goals
- Guarantee at-least-once delivery of session lifecycle events with under 5 seconds end-to-end latency for 95th percentile.
- Capture 100% of critical combat events and 30% sampled state/heartbeat payloads without exceeding the current gateway CPU budget.
- Provide schema evolution guardrails (versioning + automated validation) so new event fields can ship without client/server lock-step releases.

## Non-Goals
- Building a fully managed live-ops experimentation platform (feature flags, A/B routing) — defer until post-alpha.
- Replacing Prometheus operational metrics; gameplay analytics will complement, not supplant, infra telemetry.
- Designing long-term data warehousing beyond 6-month retention; archive/export decisions land with a future data engineering workstream.

## Proposed Architecture

```
Client SDK -> WS Envelopes -> Gateway Ingestion -> Stream Buffer -> Session Aggregator -> Column Store -> BI + Alerting
```

### Client Instrumentation
- Ship a lightweight analytics SDK in `client/` that piggybacks on the existing WebSocket connection to avoid new transports.
- SDK batches gameplay events (combat, inventory, social) and piggybacks them in a `analytics` envelope emitted alongside `input` messages at 10 Hz max.
- Provide debounce + sampling hooks so designers can tune event verbosity without code changes (configurable via JSON fetched on login).
- Fallback buffer (1 MiB cap) protects the client render loop; if overflow occurs, drop lowest-priority events and log to the debug HUD.

### Gateway Ingestion
- Extend `backend/internal/transport/ws/session.go` to accept the new `analytics` message type and enqueue events onto an internal Go channel.
- Each gateway instance hosts an `ingest.Worker` with bounded queues (default 5k events) and backpressure counters surfaced via Prometheus gauges.
- Workers publish JSON envelopes to NATS JetStream (`analytics.events`) to decouple ingestion from downstream persistence.

### Session Aggregator
- Introduce a new Go service under `backend/cmd/analytics-aggregator` that subscribes to JetStream and normalizes events.
- Responsibilities: attach server-observed metadata (latency bucket, shard, build hash), deduplicate via `(player_id, client_guid, seq)` tuple, and emit enriched records.
- Aggregator fan-outs: (1) append-only parquet batches via Apache Arrow to object storage (MinIO in dev, S3 in production), (2) Postgres-CDC staging table for near-real dashboards.

### Storage Layer
- **Hot store (≤48h):** Postgres `analytics_sessions` and `analytics_events` tables partitioned by `event_day` for quick iteration and QA queries.
- **Warm store (≤6 mo):** Daily Parquet files in `s3://telemetry/gameplay/{date}/` registered with Trino for ad-hoc analysis; lifecycle policy rolls to Glacier after retention window.
- Schema registry (JSON Schema v7) versioned in `docs/architecture/analytics-schemas/` and mirrored in code to gate deployments.

## Event Taxonomy and Sampling Strategy

| Category | Example Events | Sampling | Notes |
| --- | --- | --- | --- |
| Session Lifecycle | `session_start`, `session_end`, `session_crash`, `reconnect` | 100% | Emit at login/logout, include build, platform, shard. |
| Combat Loop | `engage_enemy`, `ability_cast`, `damage_dealt`, `damage_taken`, `death` | 100% for boss/elite, 50% otherwise | Leverage combat tags from simulation to mark encounter tier. |
| Progression & Economy | `xp_gain`, `skill_rank_up`, `loot_drop`, `inventory_move`, `craft_result` | 100% for rank-ups, 30% sampled for commodity events | Economy team needs deterministic sampling driven by player hash. |
| Social & Retention | `party_create`, `party_join`, `emote_played`, `friend_invite` | 50% | Instrument once social loops ship; ensure GDPR consent gating. |
| Heartbeat | `client_tick`, `aoi_population`, `snapshot_size` | Adaptive: start at 10% | Driven by dynamic sampler that keeps gateway <65% CPU. |

- Sampling rules live in a shared YAML (`docs/architecture/analytics-sampling.yml`) consumed by both client SDK and aggregator at boot.
- Deterministic hashing (`hash64(player_id + event_type)`) ensures the same players remain in sampled cohorts for longitudinal studies.
- Heartbeat sampling doubles automatically if combat encounters exceed 85% load thresholds for two minutes (configurable circuit breaker).

## Privacy and Anonymization Requirements
- Replace raw player identifiers with `analytics_player_id = sha256(player_uuid + salt)` where salt rotates monthly and is stored in Vault.
- Store region/country derived from IP via GeoIP, but drop the original IP after enrichment; maintain compliance with COPPA by omitting age data.
- Respect client opt-out flag (`analytics_opt_out`) and ensure events are dropped client-side and server-side when false.
- Limit PII fields: free-form chat content is excluded from analytics; only length/count metadata is recorded.
- Document data retention (6 months warm store) in `docs/governance/privacy.md` and schedule quarterly audits.

## Tooling, Dashboards, and Alerting
- Use Metabase atop Postgres + Trino for exploratory dashboards; provision templates for session funnel, combat success rates, and XP velocity.
- Wire Grafana alerts on JetStream lag, aggregator error rate, and dropped-event counters; page SRE if lag exceeds 2 minutes (sustained 5 minutes).
- Build an automated daily dbt job (running in GitHub Actions) that materializes cohort tables and pushes summary metrics to the roadmap review doc.
- Provide Looker Studio export feed for stakeholders without VPN by syncing curated views to Google Sheets via service account (read-only).

## Phased Rollout Plan
- **Phase 0 – Design Sign-off (due 2025-09-26):** Finalize schema v1.0, sampling config, and gating toggles; align with gameplay + SRE leads.
- **Phase 1 – MVP Instrumentation (2025-09-29 → 2025-10-10):** Implement client SDK stubs, gateway ingestion channel, and minimal aggregator that writes to Postgres.
- **Phase 2 – Validation & Backfill (2025-10-13 → 2025-10-24):** Run load tests via `make test-ws`, compare event counts to simulation snapshots, replay stored sessions to backfill Parquet.
- **Phase 3 – Scaling & Dashboards (2025-10-27 → 2025-11-14):** Enable JetStream retention policies, deploy Trino catalog, deliver Metabase + Grafana dashboards, and document on-call rotations.
- **Phase 4 – Hardening (2025-11-17 → 2025-12-05):** Formalize SLA, add chaos tests for message loss, and prepare ADR summarizing production readiness.

## Implementation Considerations

### Backpressure & Reliability
- Gateway queues expose Prometheus metrics (`analytics_queue_depth`, `analytics_drop_total`) and circuit breakers throttle client SDK sampling when depth >80% for 30s.
- Aggregator acknowledges JetStream messages only after persistence to Postgres succeeds; failed batches are retried with exponential backoff and DLQ (`analytics.dlq`).

### Schema Evolution
- Adopt semver for analytics schema (`analytics/v1/event.schema.json`); minors add optional fields, majors for breaking changes.
- CI adds a `make check-analytics-schema` gate that validates new schemas against captured fixtures in `docs/architecture/analytics-fixtures/`.
- Client SDK bundles the latest schema hash; gateway rejects envelopes with unknown versions and responds with `error { code: "analytics_version" }` for visibility.

### Data Quality Monitoring
- Nightly assertions compare analytics session counts vs. authenticated WebSocket sessions from operational logs; alert if variance >3%.
- Aggregator logs sample payloads to Loki with redaction rules so QA can spot malformed events.
- Introduce synthetic players in staging that emit known event sequences; regression detection catches missing fields before prod deploys.

### Integration with Existing Snapshots
- Leverage the existing `telemetry` tick (1 Hz) to piggyback snapshot size and RTT analytics without additional messages.
- Simulation engine emits combat and inventory hooks (`sim.EventBus`) to centralize server-side analytics — ensures parity if clients drop events.
- Backfill job replays stored snapshots (`docs/development/server-feature-test-plan.md` references) to synthesize analytics for historical load tests.

### Security & Compliance
- Store JetStream credentials in Vault; aggregator retrieves via existing secret loader used by the gateway (`backend/internal/config/`).
- Restrict S3 bucket with IAM policy allowing read to analytics group only; write access limited to aggregator role.
- Document data subject access request process in `docs/governance/privacy.md` during Phase 3.

## Open Questions
- Do we standardize on JetStream or evaluate Kafka before Phase 3 scale-up?
- Should designer-facing dashboards live in Metabase alone, or do we need Tableau/Looker compatibility for external reviewers?
- What budget do we assign to long-term cold storage if roadmap requires year-over-year comparisons post-alpha?

## Next Steps
- Review and approve Phase 0 deliverables during the 2025-09-24 planning sync.
- Once approved, create implementation issues for client SDK, gateway ingestion, aggregator service, and schema registry tasks.
- Coordinate with SRE/QA to extend load test harness to generate analytics payloads prior to Phase 2 validation.
