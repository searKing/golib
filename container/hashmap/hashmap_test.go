package hashmap

import "testing"

func Test(t *testing.T) {
	s := New()

	if s.Len() != 0 {
		t.Errorf("Length should be 0")
	}
	s.Remove(0)
	s.Add(5, 5)

	if s.Count() != 1 {
		t.Errorf("Length should be 1")
	}

	if !s.Contains(5) {
		t.Errorf("Contains test failed")
	}

	s.Remove(5)

	if s.Count() != 0 {
		t.Errorf("Length should be 0")
	}

	if s.Contains(5) {
		t.Errorf("The set should be empty")
	}
}
