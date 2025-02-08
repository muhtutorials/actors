package cluster

type MemberSet struct {
	members map[string]*Member
}

func NewMemberSet(members ...*Member) *MemberSet {
	m := make(map[string]*Member)
	for _, member := range members {
		m[member.ID] = member
	}
	return &MemberSet{
		members: m,
	}
}

func (s *MemberSet) Len() int {
	return len(s.members)
}

func (s *MemberSet) GetByHost(host string) *Member {
	var member *Member
	for _, m := range s.members {
		if m.Host == host {
			member = m
		}
	}
	return member
}

func (s *MemberSet) Add(m *Member) {
	s.members[m.ID] = m
}

func (s *MemberSet) Contains(m *Member) bool {
	_, ok := s.members[m.ID]
	return ok
}

func (s *MemberSet) Remove(m *Member) {
	delete(s.members, m.ID)
}

func (s *MemberSet) RemoveByHost(host string) {
	member := s.GetByHost(host)
	if member != nil {
		s.Remove(member)
	}
}

func (s *MemberSet) Slice() []*Member {
	members := make([]*Member, len(s.members))
	i := 0
	for _, member := range s.members {
		members[i] = member
		i++
	}
	return members
}

func (s *MemberSet) ForEach(fn func(*Member) bool) {
	for _, member := range s.members {
		if !fn(member) {
			break
		}
	}
}

func (s *MemberSet) FilterByKind(kind string) []*Member {
	var members []*Member
	for _, member := range s.members {
		if member.HasKind(kind) {
			members = append(members, member)
		}
	}
	return members
}

// Difference calculates the difference of sets "s.members" and "members".
func (s *MemberSet) Difference(members []*Member) []*Member {
	var (
		diff []*Member
		m    = make(map[string]*Member)
	)
	for _, member := range members {
		m[member.ID] = member
	}
	for _, member := range s.members {
		if _, ok := m[member.ID]; !ok {
			diff = append(diff, member)
		}
	}
	return diff
}
