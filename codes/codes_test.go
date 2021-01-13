package codes

import "testing"

func TestAreCodesValid(t *testing.T) {
	type args struct {
		countryCode string
		stateCode   string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
		{"France ok", args{"FR", "04"}, true},
		{"France ko", args{"FR", "00"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AreCodesValid(tt.args.countryCode, tt.args.stateCode); got != tt.want {
				t.Errorf("AreCodesValid() = %v, want %v", got, tt.want)
			}
		})
	}
}
