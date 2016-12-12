package requestid

import (
	"regexp"
	"testing"
)

func TestUUID(t *testing.T) {
	u := NewUUID()
	if len(u) != 16 {
		t.Errorf("Invalid uuid: %s", u.String())
	}
	m, err := regexp.MatchString(`([a-f\d]{8}(-[a-f\d]{4}){3}-[a-f\d]{12}?)`, u.String())
	if err != nil {
		t.Error(err)
	}
	if !m {
		t.Errorf("%s is not valid format", u.String())
	}
}

func BenchmarkUUID(b *testing.B) {
	b.Logf("Benchmarking UUID")
	m := make(map[string]int, 1000)
	for i := 0; i < b.N; i++ {
		u := NewUUID()
		b.StopTimer()
		c := m[u.String()]
		if c > 0 {
			b.Fatalf("Duplicate uuid %s count %d", u, c)
		}
		m[u.String()] = c + 1
		b.StartTimer()
	}
}
