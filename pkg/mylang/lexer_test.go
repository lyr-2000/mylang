package mylang

import "testing"

func TestTrimComment(t *testing.T) {
	t.Run("trim-comment:1",func(t *testing.T) {
		code := "abc{comment}\ndef"
		got := TrimComment(code)
		want := "abc\ndef"
		if got != want {
			t.Errorf("TrimComment() = %v, want %v", got, want)
		}
	})
}
