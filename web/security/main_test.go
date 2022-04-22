package security

import "testing"

func TestTokenUniqness(t *testing.T) {
	t1 := GenerateToken()
	t2 := GenerateToken()
	if t1 == t2 {
		t.Log("Tokens are the same:", t1, t2)
		t.Fail()
	}
}

func BenchmarkTokenCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = GenerateToken()
	}
}
