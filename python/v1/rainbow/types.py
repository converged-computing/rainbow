from dataclasses import dataclass


@dataclass
class SatisfyResponse:
    clusters: list
    cluster: str
    total_matches: int
    total_mismatches: int
    total_clusters: int
    status: int
