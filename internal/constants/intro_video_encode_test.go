package constants

import "testing"

func TestDefaultIntroVideoCompressPreset(t *testing.T) {
	if DefaultIntroVideoCompressPreset() != IntroVideoCompressPresetFine {
		t.Fatalf("DefaultIntroVideoCompressPreset() = %q, want %q", DefaultIntroVideoCompressPreset(), IntroVideoCompressPresetFine)
	}
}

func TestValidIntroVideoCompressPreset(t *testing.T) {
	cases := []struct {
		value string
		want  bool
	}{
		{"very_low", true},
		{"medium", true},
		{"fine", true},
		{"high", true},
		{"original", true},
		{"", false},
		{"invalid", false},
	}
	for _, tc := range cases {
		if got := ValidIntroVideoCompressPreset(tc.value); got != tc.want {
			t.Fatalf("ValidIntroVideoCompressPreset(%q) = %v, want %v", tc.value, got, tc.want)
		}
	}
}

func TestIntroVideoEncodeSettingsForPreset(t *testing.T) {
	fine := IntroVideoEncodeSettingsForPreset(IntroVideoCompressPresetFine)
	if fine.CRF != "22" || fine.MaxWidth != 1920 || fine.AudioBitrate != "160k" || fine.TimeoutSecs != 300 {
		t.Fatalf("unexpected fine settings: %+v", fine)
	}

	high := IntroVideoEncodeSettingsForPreset(IntroVideoCompressPresetHigh)
	if high.Preset != "slow" || high.TimeoutSecs != 600 {
		t.Fatalf("unexpected high settings: %+v", high)
	}

	original := IntroVideoEncodeSettingsForPreset(IntroVideoCompressPresetOriginal)
	if !original.SkipCompress {
		t.Fatal("expected original preset to skip compression")
	}
}

func TestIntroVideoEncodeSettingsForPresetValueFallback(t *testing.T) {
	settings := IntroVideoEncodeSettingsForPresetValue("not-a-preset")
	fine := IntroVideoEncodeSettingsForPreset(IntroVideoCompressPresetFine)
	if settings != fine {
		t.Fatalf("unexpected fallback settings: %+v", settings)
	}
}

func TestIntroVideoShouldCompress(t *testing.T) {
	fine := IntroVideoEncodeSettingsForPreset(IntroVideoCompressPresetFine)
	original := IntroVideoEncodeSettingsForPreset(IntroVideoCompressPresetOriginal)

	cases := []struct {
		name     string
		fileSize int64
		settings IntroVideoEncodeSettings
		want     bool
	}{
		{"below threshold compresses with fine preset", IntroVideoCompressMinBytes - 1, fine, false},
		{"at threshold compresses", IntroVideoCompressMinBytes, fine, true},
		{"above threshold compresses", IntroVideoCompressMinBytes + 1, fine, true},
		{"skip compress preset", IntroVideoCompressMinBytes + 1, original, false},
		{"small file with skip compress", 1, original, false},
	}
	for _, tc := range cases {
		if got := IntroVideoShouldCompress(tc.fileSize, tc.settings); got != tc.want {
			t.Fatalf("%s: IntroVideoShouldCompress(%d, ...) = %v, want %v", tc.name, tc.fileSize, got, tc.want)
		}
	}
}
