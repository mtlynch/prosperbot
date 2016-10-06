package notes

import (
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/mtlynch/gofn-prosper/prosper"
)

type mockNoteFetcher struct {
	calls int
	notes []prosper.Note
	err   error
}

func (ls *mockNoteFetcher) Notes(p prosper.NotesParams) (prosper.NotesResponse, error) {
	ls.calls++
	actualLimit := p.Offset + p.Limit
	if actualLimit > len(ls.notes) {
		actualLimit = len(ls.notes)
	}
	result := ls.notes[p.Offset:actualLimit]
	return prosper.NotesResponse{
		Result:      result,
		ResultCount: len(result),
		TotalCount:  len(ls.notes),
	}, ls.err
}

func makeNotes(count int) []prosper.Note {
	notes := []prosper.Note{}
	for i := 0; i < count; i++ {
		notes = append(notes, prosper.Note{ListingNumber: prosper.ListingNumber(i)})
	}
	return notes
}

type ByListingNumber []prosper.Note

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
		serverNotes []prosper.Note
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
		notes := make(chan prosper.Note)
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
		var gotNotes []prosper.Note
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
