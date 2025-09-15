# 0001: Local sharding for world cells

- **Status**: Accepted
- **Context**: Early prototypes needed a simple way to scale the world and limit broadcast scope.
- **Decision**: Partition the world into grid cells and let a single simulation process manage all cells locally. Players hand over between cells with a 3×3 area of interest.
- **Consequences**: Simplifies MVP implementation and AOI queries, but requires new infrastructure when moving to multi-node sharding.

