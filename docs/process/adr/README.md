# Architecture Decision Records (ADRs)

This directory contains architectural decision records for the prototype-game project.

## Template and Guidelines

- [TEMPLATE.md](TEMPLATE.md) - Standard template for new ADRs
- [PEER_REVIEW_GUIDELINES.md](PEER_REVIEW_GUIDELINES.md) - Guidelines for reviewing ADRs

## Recorded Decisions

### ADR-0001: Local Sharding
- [0001-local-sharding.md](0001-local-sharding.md) - Core decision on local sharding approach
- [0001-local-sharding-peer-review.md](0001-local-sharding-peer-review.md) - Peer review
- [0001-local-sharding-review-summary.md](0001-local-sharding-review-summary.md) - Review summary

### ADR-0002: Cross-Node Handover Protocol
- [0002-cross-node-handover-protocol.md](0002-cross-node-handover-protocol.md) - Handover protocol design
- [0002-cross-node-handover-protocol-REVIEW.md](0002-cross-node-handover-protocol-REVIEW.md) - Review

### ADR-0003: Distributed Cell Assignment
- [0003-distributed-cell-assignment.md](0003-distributed-cell-assignment.md) - Cell assignment strategy
- [0003-distributed-cell-assignment-REVIEW.md](0003-distributed-cell-assignment-REVIEW.md) - Review
- [0003-implementation-guide.md](0003-implementation-guide.md) - Implementation guidance
- [0003-peer-review-summary.md](0003-peer-review-summary.md) - Peer review summary

## Process

All architectural decisions should be documented using the ADR process:

1. Use [TEMPLATE.md](TEMPLATE.md) for new decisions
2. Follow [PEER_REVIEW_GUIDELINES.md](PEER_REVIEW_GUIDELINES.md) for reviews
3. Number ADRs sequentially (ADR-0001, ADR-0002, etc.)
4. Update this README when adding new ADRs