package test

import (
	"../server"
	"testing"
)

func BenchmarkParsingStats(b *testing.B) {
	one := []byte{0, 0, 0, 1, 0, 0, 0, 123, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 3, 232, 91, 207, 227, 187}
	//fmt.Println(one)
	b.ResetTimer()
	content := new(server.ModuleStats)
	for i := 0; i < b.N; i++ {
		server.ByteToModule(one[0:4], content, 1)
		server.ByteToModule(one[4:8], content, 2)
		server.ByteToModule(one[8:9], content, 3)
		server.ByteToModule(one[9:13], content, 4)
		server.ByteToModule(one[13:17], content, 5)
		server.ByteToModule(one[17:21], content, 6)
		server.ByteToModule(one[21:25], content, 7)
	}
}

func BenchmarkAll(b *testing.B) {
	one := []byte{0, 0, 0, 1, 0, 0, 0, 123, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 3, 232, 91, 207, 227, 187}
	//fmt.Println(one)
	s := server.StatsServer{}
	s.TimeInterval = 5
	s.TimeKeyInterval = 5
	b.ResetTimer()
	content := new(server.ModuleStats)
	for i := 0; i < b.N; i++ {
		server.ByteToModule(one[0:4], content, 1)
		server.ByteToModule(one[4:8], content, 2)
		server.ByteToModule(one[8:9], content, 3)
		server.ByteToModule(one[9:13], content, 4)
		server.ByteToModule(one[13:17], content, 5)
		server.ByteToModule(one[17:21], content, 6)
		server.ByteToModule(one[21:25], content, 7)
		s.CalculateModule(content)
	}
}
