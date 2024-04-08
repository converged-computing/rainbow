package constraint

/*
- priority: 1
 steps:
 - filter: "nodes_free > 0"
 - calc: "build_cost=(cost_per_node_hour * (memory_gb_per_node * seconds_per_gb)/60/60))"
 - sort_descending: build_cost
 - select: random
- priority: 2
  steps:
  - filter: "nodes_free > 0"
  - calc: "memory_min=min(100, memory_gb_per_node - 100)"
  - calc: "build_cost=(cost_per_node_hour * (memory_min * seconds_per_gb)/60/60))"
  - sort_descending: build_cost
  - select: random
*/

type ConstraintPriority struct {
	Priority int32               `yaml:"priority"`
	Steps    []map[string]string `yaml:"steps"`
}
