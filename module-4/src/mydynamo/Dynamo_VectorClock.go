package mydynamo

type VectorClock struct {
	//todo
	TimeList map[string]int
}

//Creates a new VectorClock
func NewVectorClock() VectorClock {
	timelist := make(map[string]int)
	return VectorClock{
			TimeList: timelist,
	}
}

//Returns true if the other VectorClock is causally descended from this one
func (s VectorClock) LessThan(otherClock VectorClock) bool {
	if s.Equals(otherClock) {
		return false
	}
	for nodeId, v := range s.TimeList {
		if val, ok := otherClock.TimeList[nodeId]; ok {
			//if the other VectorClock has a version associated with this id
			if v > val {
				return false
		}
	}else{
		//if no version associated with this id in the other VectorClock,we treat it
		//as a zero value, so it suggests that our version is greater than the otherClock
		return false
		}
	}
	return true
}

//Returns true if neither VectorClock is causally descended from the other
func (s VectorClock) Concurrent(otherClock VectorClock) bool {
	if s.Equals(otherClock) || s.LessThan(otherClock) || otherClock.LessThan(s) {
		return false
	}
	return true
}

//Increments this VectorClock at the element associated with nodeId
func (s *VectorClock) Increment(nodeId string) {
	s.TimeList[nodeId] += 1
}

func Max(x, y int) int {
    if x < y {
        return y
    }
    return x
}

//Changes this VectorClock to be causally descended from all VectorClocks in clocks
func (s *VectorClock) Combine(clocks []VectorClock) {
	newClock := NewVectorClock()
	for _, clk := range clocks {
		for nodeId, v1 := range clk.TimeList {
			if v2, ok := newClock.TimeList[nodeId]; ok {
				newClock.TimeList[nodeId] = Max(v1, v2)
			} else {
				newClock.TimeList[nodeId] = v1
			}
		}
	}

	for nodeId, v1 := range newClock.TimeList {
		if v2, ok := s.TimeList[nodeId]; ok {
			s.TimeList[nodeId] = Max(v1, v2)
		} else {
			s.TimeList[nodeId] = v1
		}
	}
}

//Tests if two VectorClocks are equal
func (s *VectorClock) Equals(otherClock VectorClock) bool {
	if len(s.TimeList) != len(otherClock.TimeList) {
        return false
    }
	for nodeId, v1 := range s.TimeList {
		if v2, ok := otherClock.TimeList[nodeId]; !ok || v2 != v1 {
            return false
        }
	}
	return true
}
