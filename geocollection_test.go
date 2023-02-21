// Copyright 2023 SpotHero
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package geocollection

import (
	"testing"

	"github.com/golang/geo/s2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testCell = struct {
	cellID   s2.CellID
	lat, lon float64
}

// cell IDs were generated using the Sidewalk Labs s2 demo
// https://s2.sidewalklabs.com/regioncoverer/
var (
	// downtown Chicago
	cell1 = testCell{
		cellID: s2.CellIDFromToken("880e2cbc904a0c29"),
		lat:    41.87963549397698,
		lon:    -87.63028184499035,
	}
	// midtown Manhattan
	cell2 = testCell{
		cellID: s2.CellIDFromToken("89c25900437"),
		lat:    40.75306726395187,
		lon:    -73.98119781456353,
	}
)

type testItem struct {
	key      int
	contents string
	lat, lon float64
}

func TestCollection_Set(t *testing.T) {
	type cellContains struct {
		cellID s2.CellID
		item   testItem
	}
	tests := []struct {
		name                   string
		items                  []testItem
		expectedCellIDContains []cellContains
	}{
		{
			name:                   "Should set an item",
			items:                  []testItem{{0, "0", cell1.lat, cell1.lon}},
			expectedCellIDContains: []cellContains{{cell1.cellID, testItem{0, "0", cell1.lat, cell1.lon}}},
		}, {
			name: "Should set multiple items",
			items: []testItem{
				{0, "0", cell1.lat, cell1.lon},
				{1, "1", cell2.lat, cell2.lon},
			},
			expectedCellIDContains: []cellContains{
				{cell1.cellID, testItem{0, "0", cell1.lat, cell1.lon}},
				{cell2.cellID, testItem{1, "1", cell2.lat, cell2.lon}}},
		}, {
			name: "Should replace an item's coordinates",
			items: []testItem{
				{0, "0", cell1.lat, cell1.lon},
				{0, "0", cell2.lat, cell2.lon},
			},
			expectedCellIDContains: []cellContains{{cell2.cellID, testItem{0, "0", cell2.lat, cell2.lon}}},
		}, {
			name: "Should replace an item's contents only",
			items: []testItem{
				{0, "0", cell1.lat, cell1.lon},
				{0, "1", cell1.lat, cell1.lon},
			},
			expectedCellIDContains: []cellContains{{cell1.cellID, testItem{0, "1", cell1.lat, cell1.lon}}},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cl := NewCollection()
			for _, item := range test.items {
				cl.Set(item.key, item.contents, item.lat, item.lon)
			}
			assert.Len(t, cl.keys, len(test.expectedCellIDContains))
			// assert that the location's cell has been cached at every cell level (31 of them)
			assert.Len(t, cl.cells, 31)
			for _, expectedContains := range test.expectedCellIDContains {
				expectedCellID := expectedContains.cellID
				assert.Contains(t, cl.keys, expectedContains.item.key)
				require.Contains(t, cl.cells[expectedCellID.Level()][expectedCellID.Pos()], expectedContains.item.key)
				assert.Contains(t, cl.cells[expectedCellID.Level()], expectedCellID.Pos())
				require.Contains(t, cl.items, expectedContains.item.key)
				assert.Equal(
					t,
					cl.items[expectedContains.item.key],
					collectionContents{
						contents:  expectedContains.item.contents,
						latitude:  expectedContains.item.lat,
						longitude: expectedContains.item.lon,
					},
				)
			}
		})
	}
}

func TestCollection_Delete(t *testing.T) {
	cell := cell1
	item := testItem{key: 0, lat: cell.lat, lon: cell.lon}
	tests := []struct {
		name                  string
		deleteKey             int
		expectedRemainingKeys []int
	}{
		{
			name:                  "Should delete an item",
			deleteKey:             item.key,
			expectedRemainingKeys: []int{},
		}, {
			name:                  "Deleting an item that does not exist should be ok",
			deleteKey:             item.key + 1,
			expectedRemainingKeys: []int{item.key},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cl := NewCollection()
			cl.Set(item.key, item.contents, item.lat, item.lon)
			cl.Delete(test.deleteKey)
			assert.NotContains(t, cl.keys, test.deleteKey)
			for level := maxCellLevel; level >= 0; level-- {
				assert.NotContains(t, cl.cells[level][cell.cellID.Pos()], test.deleteKey)
				for _, remainingID := range test.expectedRemainingKeys {
					assert.Contains(t, cl.cells[level][cell.cellID.Parent(level).Pos()], remainingID)
				}
			}
			for _, remainingID := range test.expectedRemainingKeys {
				assert.Contains(t, cl.keys, remainingID)
			}
		})
	}
}

func TestCollection_ItemsWithinDistance(t *testing.T) {
	item1 := testItem{key: 0, contents: "1", lat: cell1.lat, lon: cell1.lon}
	item2 := testItem{key: 1, contents: "2", lat: cell2.lat, lon: cell2.lon}
	tests := []struct {
		name                                 string
		searchLat, searchLon, distanceMeters float64
		useFastAlgorithm                     bool
		expectedContents                     []string
	}{
		{
			name:             "Search should return relevant results",
			searchLat:        cell1.lat - 0.001,
			searchLon:        cell1.lon - 0.001,
			distanceMeters:   1000,
			useFastAlgorithm: false,
			expectedContents: []string{item1.contents},
		},
		{
			name:             "Search should return relevant with the fast cover algorithm",
			searchLat:        cell1.lat - 0.001,
			searchLon:        cell1.lon - 0.001,
			distanceMeters:   1000,
			useFastAlgorithm: true,
			expectedContents: []string{item1.contents},
		}, {
			name:             "Search should return multiple relevant results",
			searchLat:        cell1.lat,
			searchLon:        cell1.lon,
			distanceMeters:   4000000,
			useFastAlgorithm: false,
			expectedContents: []string{item1.contents, item2.contents},
		}, {
			name:             "Search should return no results when no items are close by",
			searchLat:        0,
			searchLon:        0,
			distanceMeters:   1,
			useFastAlgorithm: false,
			expectedContents: []string{},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cl := NewCollection()
			cl.Set(item1.key, item1.contents, item1.lat, item1.lon)
			cl.Set(item2.key, item2.contents, item2.lat, item2.lon)
			results, _ := cl.ItemsWithinDistance(
				test.searchLat, test.searchLon, test.distanceMeters, SearchCoveringParameters{
					MaxLevel: 5, MinLevel: 5, LevelMod: 1, MaxCells: 5, UseFastCovering: test.useFastAlgorithm})
			assert.Len(t, results, len(test.expectedContents))
			for _, content := range test.expectedContents {
				assert.Contains(t, results, content)
			}
		})
	}
}

func TestCollection_ItemByKey(t *testing.T) {
	c := NewCollection()
	c.items[1] = collectionContents{contents: "1"}
	tests := []struct {
		name             string
		key              interface{}
		expectedContents interface{}
	}{
		{"Item is retrieved from collection by its key", 1, "1"},
		{"No key exists returns nil", 2, nil},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expectedContents, c.ItemByKey(test.key))
		})
	}
}

func TestEarthDistanceMeters(t *testing.T) {
	// pick 2 points off a map that are roughly 105 meters of each other
	p1 := NewPointFromLatLng(41.883170, -87.632278)
	p2 := NewPointFromLatLng(41.883178, -87.630916)
	assert.InDelta(t, 105, EarthDistanceMeters(p1, p2), 10)
}
