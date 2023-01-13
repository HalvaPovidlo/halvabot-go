package item

import "testing"

func TestGetIDFromURL(t *testing.T) {
	type test struct {
		in  string
		out string
	}

	testCases := []test{
		{
			in:  "https://www.youtube.com/watch?v=hDfFXWinkAk",
			out: "youtube_hDfFXWinkAk",
		},
		{
			in:  "https://youtube.com/watch?v=hDfFXWinkAk",
			out: "youtube_hDfFXWinkAk",
		},
	}

	for i := range testCases {
		tc := &testCases[i]
		id := GetIDFromURL(tc.in).String()
		if id != tc.out {
			t.Errorf("input: %sgot %q, wanted %q", tc.in, id, tc.out)
		}
	}
}

func TestTestYoutubeURL(t *testing.T) {
	type test struct {
		in  string
		out bool
	}

	testCases := []test{
		{
			in:  "https://www.youtube.com/watch?v=hDfFXWinkAk",
			out: true,
		},
		{
			in:  "https://youtube.com/watch?v=hDfFXWinkAk",
			out: true,
		},
		{
			in:  "httssps://www.youtube.com/watch?v=hDfFXWinkAk",
			out: false,
		},
		{
			in:  "https://vk.com/watch?v=hDfFXWinkAk",
			out: false,
		},
	}

	for i := range testCases {
		tc := &testCases[i]
		ok := TestYoutubeURL(tc.in)
		if ok != tc.out {
			t.Errorf("input: %s got %t, wanted %t", tc.in, ok, tc.out)
		}
	}
}
