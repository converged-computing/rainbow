package types

import (
	"fmt"
)

// Resource needs can be used for:

//  1. a slot, in which case we use the Found/Needed fields, and lookups buy
//     many types.
//  2. A single resource type, in which case we define Type and
//     ignore Found/Needed
type MatchAlgorithmNeeds map[string]map[string]bool

// Serialize slot resource needs into a struct that is easier to parse
type ResourceNeeds struct {
	SubsystemSatisfied bool
	ResourceSatisfied  bool

	// Lookup by vertex type
	// type -> subsystem -> attribute -> isSatisfied
	Subsystems map[string]MatchAlgorithmNeeds
	Resources  map[string]int32

	// Needed vs found
	Found  int32
	Needed int32

	// Only needed of the ResourceNeeds is for a single type of resource
	Type string

	// Caches to use for reset
	SubsystemsOriginal map[string]MatchAlgorithmNeeds
	ResourcesOriginal  map[string]int32
}

func (s *ResourceNeeds) Reset() {
	s.Subsystems = s.SubsystemsOriginal
	s.Resources = s.ResourcesOriginal
	s.SubsystemSatisfied = false
	s.ResourceSatisfied = false
}

func (s *ResourceNeeds) SummarizeRemaining() string {
	summary := ""
	for resourceType, count := range s.Resources {
		if count > 0 {
			summary += fmt.Sprintf(": %s=%d", resourceType, count)
			// Also get number of subsystem needs
			sNeeds, ok := s.Subsystems[resourceType]
			if ok {
				for subsystem, attributes := range sNeeds {
					count = 0
					for _, isSatsified := range attributes {
						if !isSatsified {
							count += 1
						}
					}
					if count > 0 {
						summary += fmt.Sprintf(" %s=%d", subsystem, count)
					}
				}
			}
		}
	}
	return summary
}

func (s *ResourceNeeds) Satisfied() bool {
	return s.Found >= s.Needed
}

func (s *ResourceNeeds) AreResourcesSatisfied() bool {
	if s.ResourceSatisfied {
		return true
	}
	for _, count := range s.Resources {
		if count > 0 {
			return false
		}
	}
	s.ResourceSatisfied = true
	return true
}

func (s *ResourceNeeds) AreSubsystemsSatisfied() bool {
	if s.SubsystemSatisfied {
		return true
	}
	for _, needs := range s.Subsystems {
		for _, typeNeeds := range needs {
			for _, isSatisfied := range typeNeeds {
				if !isSatisfied {
					return false
				}
			}
		}
	}
	// Cache result so we don't redo big loops
	s.SubsystemSatisfied = true
	return true
}

func (s *ResourceNeeds) AllSatisfied() bool {
	return s.AreSubsystemsSatisfied() && s.AreResourcesSatisfied()
}
