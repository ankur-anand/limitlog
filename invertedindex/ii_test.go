package invertedindex_test

import (
	"reflect"
	"testing"

	"github.com/ankur-anand/limitlog/invertedindex"
)

func TestInvertedIndex_Add_Delete_Search(t *testing.T) {
	ii := invertedindex.NewInvertedIndex()

	ii.Add("We need to manage logs on a system with limited memory.", 1)
	ii.Add("We need to query which of the logs contain a given word.", 2)
	ii.Add("The first line of the input is the maximum size of logs you should keep S.", 3)
	ii.Add("The last line contains the single word END denoting the end of the program.", 4)

	tCases := []struct {
		name         string
		wordToSearch string
		resultDocsID map[int64]struct{}
	}{
		{
			name:         "searhcing we",
			wordToSearch: "We",
			resultDocsID: map[int64]struct{}{
				1: {},
				2: {},
			},
		},
		{
			name:         "searhcing logs",
			wordToSearch: "logs",
			resultDocsID: map[int64]struct{}{
				1: {},
				2: {},
				3: {},
			},
		},
		{
			name:         "searhcing the",
			wordToSearch: "the",
			resultDocsID: map[int64]struct{}{
				2: {},
				3: {},
				4: {},
			},
		},
		{
			name:         "searhcing end",
			wordToSearch: "end",
			resultDocsID: map[int64]struct{}{
				4: {},
			},
		},
	}

	for _, tc := range tCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ii.Search(tc.wordToSearch)
			if !reflect.DeepEqual(result, tc.resultDocsID) {
				t.Errorf("expected index to document map didn't match")
			}
		})
	}

	ii.RemoveAssociation("The last line contains the single word END denoting the end of the program.", 4)
	res := ii.Search("end")
	if len(res) != 0 {
		t.Errorf("deleted document id should not be present.")
	}
	ii.RemoveAssociation("We need to query which of the logs contain a given word.", 2)
	res = ii.Search("we")
	if !reflect.DeepEqual(res, map[int64]struct{}{
		1: {},
	}) {
		t.Errorf("deleted document id should not be present.")
	}
}
