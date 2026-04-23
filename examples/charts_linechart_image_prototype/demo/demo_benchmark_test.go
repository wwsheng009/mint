package demo

import "testing"

func BenchmarkPrototypeLineChartTextPaint(b *testing.B) {
	inst := prototypeChartInstance(0)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = inst.Paint(0, 0)
	}
}

func BenchmarkPrototypeLineChartRequestedImageBackendPaint(b *testing.B) {
	inst := prototypeChartInstance(1)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = inst.Paint(0, 0)
	}
}
