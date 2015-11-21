package guid

import (
	"testing"
)

func BenchmarkGUID(b *testing.B) {
	factory := &GuidFactory{}
	for i := 0; i < b.N; i++ {
		guid, err := factory.NewGUID(0)
		if err != nil {
			b.FailNow()
		}
		println(guid)
	}
}
