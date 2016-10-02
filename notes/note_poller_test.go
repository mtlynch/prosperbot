package notes

import (
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/mtlynch/gofn-prosper/types"
)

type mockNoteFetcher struct {
	calls int
	notes []types.Note
	err   error
}

func (ls *mockNoteFetcher) Notes(offset, limit int) (types.NotesResponse, error) {
	ls.calls++
	actualLimit := offset + limit
	if actualLimit > len(ls.notes) {
		actualLimit = len(ls.notes)
	}
	result := ls.notes[offset:actualLimit]
	return types.NotesResponse{
		Result:      result,
		ResultCount: len(result),
		TotalCount:  len(ls.notes),
	}, ls.err
}

func makeNotes(count int) []types.Note {
	notes := []types.Note{}
	for i := 0; i < count; i++ {
		notes = append(notes, types.Note{ListingNumber: types.ListingNumber(i)})
	}
	return notes
}

type ByListingNumber []types.Note

func (s ByListingNumber) Len() int {
	return len(s)
}

func (s ByListingNumber) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s ByListingNumber) Less(i, j int) bool {
	return s[i].ListingNumber < s[j].ListingNumber
}

func TestNotePoller(t *testing.T) {
	var tests = []struct {
		serverNotes []types.Note
		searchErr   error
		wantCalls   int
	}{
		{
			serverNotes: makeNotes(1),
			wantCalls:   1,
		},
		{
			serverNotes: makeNotes(2),
			wantCalls:   1,
		},
		{
			serverNotes: makeNotes(25),
			wantCalls:   1,
		},
		{
			serverNotes: makeNotes(26),
			wantCalls:   2,
		},
		{
			serverNotes: makeNotes(50),
			wantCalls:   2,
		},
		{
			serverNotes: makeNotes(51),
			wantCalls:   3,
		},
	}
	for _, tt := range tests {
		notes := make(chan types.Note)
		noteFetcher := mockNoteFetcher{
			notes: tt.serverNotes,
			err:   tt.searchErr,
		}
		notePoller := notePoller{
			nf:           &noteFetcher,
			notes:        notes,
			pollInterval: 10 * time.Second,
		}
		go notePoller.Run()
		var gotNotes []types.Note
		for i := 0; i < len(tt.serverNotes); i++ {
			gotNotes = append(gotNotes, <-notes)
		}
		sort.Sort(ByListingNumber(gotNotes))
		if !reflect.DeepEqual(tt.serverNotes, gotNotes) {
			t.Fatalf("for notes size %d, unexpected server notes. got: %+v, want: %+v", len(tt.serverNotes), gotNotes, tt.serverNotes)
		}
		if noteFetcher.calls != tt.wantCalls {
			t.Errorf("for notes size %d, unexpected calls to client.Notes. got: %d, want: %d", len(tt.serverNotes), noteFetcher.calls, tt.wantCalls)
		}
	}
}
