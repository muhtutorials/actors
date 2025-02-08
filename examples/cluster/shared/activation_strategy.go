package shared

import (
	"github.com/muhtutorials/actors/cluster"
	"math/rand"
)

type RegionBasedActivationStrategy struct {
	fallback string
}

func NewRegionBasedActivationStrategy(fallback string) RegionBasedActivationStrategy {
	return RegionBasedActivationStrategy{fallback: fallback}
}

func (as *RegionBasedActivationStrategy) ActivateOnMember(details cluster.ActivationDetails) *cluster.Member {
	members := filterMembersByRegion(details.Members, details.Region)
	if len(members) > 0 {
		return members[rand.Intn(len(members))]
	}
	// if we could not find a member for the region, try to fall back.
	members = filterMembersByRegion(details.Members, as.fallback)
	if len(members) > 0 {
		return members[rand.Intn(len(members))]
	}
	return nil
}

func filterMembersByRegion(members []*cluster.Member, region string) []*cluster.Member {
	filteredMembers := make([]*cluster.Member, 0)
	for _, member := range members {
		if member.Region == region {
			filteredMembers = append(filteredMembers, member)
		}
	}
	return filteredMembers
}
