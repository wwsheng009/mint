package candlestick

import "github.com/wwsheng009/mint/runtime/style"

// Candle describes one OHLC candle.
type Candle struct {
	Label  string
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume float64
	Style  style.Style
}

func copyCandleSlice(src []Candle) []Candle {
	if len(src) == 0 {
		return nil
	}
	dst := make([]Candle, len(src))
	copy(dst, src)
	return dst
}

func candleSlicesEqual(a, b []Candle) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
