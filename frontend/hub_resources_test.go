package frontend

import "testing"

func TestHubFilePreviewKind(t *testing.T) {
	cases := []struct {
		ext  string
		want HubFilePreviewKind
	}{
		{ext: ".PNG", want: HubFilePreviewKindImage},
		{ext: "jpeg", want: HubFilePreviewKindImage},
		{ext: "pdf", want: HubFilePreviewKindPDF},
		{ext: "mp4", want: HubFilePreviewKindVideo},
		{ext: "docx", want: HubFilePreviewKindDocument},
	}
	for _, tc := range cases {
		if got := hubFilePreviewKind(tc.ext); got != tc.want {
			t.Fatalf("hubFilePreviewKind(%q) = %q, want %q", tc.ext, got, tc.want)
		}
	}
}
