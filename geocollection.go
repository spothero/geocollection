// Copyright 2021 SpotHero
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

// Package geocollection provides a data structure for storing and quickly
// searching for items based on geographic coordinates on Earth.
package geocollection

import (
	"sync"

	"github.com/golang/geo/s1"
	"github.com/golang/geo/s2"
)

// EarthRadiusMeters is an approximate representation of the earth's radius in meters.
const EarthRadiusMeters = 6371008.8

// maxCellLevel is the number of levels to specify a leaf cell in s2 -- this is copied
// from s2 because they do not export this value
const maxCellLevel = 30

// cellItems is a map of cell ids to the set of keys pertaining to items geographically contained in that cell
type cellItems map[uint64]map[interface{}]bool

// itemIndex keeps track of which cells a given item belongs to in order to enable fast deletions
type itemIndex struct {
	cellPosition uint64
	cellLevel    int
}

// collectionContents stores the contents of a key and the original latitude and longitude
// stored with the key.
type collectionContents struct {
	contents            interface{}
	latitude, longitude float64
}

// Collection implements the GeoLocationCollection interface and provides a location based
// cache
type Collection struct {
	// cells is a map of cell level to the items contained in each cell at that zoom level
	cells map[int]cellItems
	// keys maps each key stored to its associated cells to enable fast deletions
	keys map[interface{}][]itemIndex
	// items maps the item key to the item contents
	items map[interface{}]collectionContents
	mutex *sync.RWMutex
}

// LocationCollection defines the interface for interacting with Geo-based collections
type LocationCollection interface {
	Set(key, contents interface{}, latitude, longitude float64)
	Delete(key interface{})
	ItemsWithinDistance(latitude, longitude, distanceMeters float64, params SearchCoveringParameters) ([]interface{}, SearchCoveringResult)
	ItemByKey(key interface{}) interface{}
}

// NewCollection creates a new collection
func NewCollection() Collection {
	return Collection{
		cells: make(map[int]cellItems),
		keys:  make(map[interface{}][]itemIndex),
		items: make(map[interface{}]collectionContents),
		mutex: &sync.RWMutex{},
	}
}

// Set adds an item with a given key to the geo collection at a particular latitude and longitude.
// If the given key already exists in the collection, it is created, otherwise the contents and location is
// updated to the new values.
func (c Collection) Set(key, contents interface{}, latitude, longitude float64) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	newContents := collectionContents{contents: contents, latitude: latitude, longitude: longitude}
	if existingContents, ok := c.items[key]; ok &&
		existingContents.latitude == latitude && existingContents.longitude == longitude {
		// contents changed but the location has not, swap contents and exit
		c.items[key] = newContents
		return
	}

	c.delete(key)
	c.items[key] = newContents
	c.keys[key] = make([]itemIndex, 0, maxCellLevel)
	leafCellID := s2.CellIDFromLatLng(s2.LatLngFromDegrees(latitude, longitude))
	for level := maxCellLevel; level >= 0; level-- {
		if _, ok := c.cells[level]; !ok {
			c.cells[level] = make(cellItems)
		}
		cellPos := leafCellID.Parent(level).Pos()
		if _, ok := c.cells[level][cellPos]; !ok {
			c.cells[level][cellPos] = make(map[interface{}]bool)
		}
		c.cells[level][cellPos][key] = true
		c.keys[key] = append(
			c.keys[key],
			itemIndex{
				cellPosition: cellPos,
				cellLevel:    level,
			},
		)
	}
}

// Delete removes an item by its key from the collection.
func (c Collection) Delete(key interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.delete(key)
}

// delete is the internal function that actually performs the deletion.
func (c Collection) delete(key interface{}) {
	delete(c.items, key)
	itemIndices, ok := c.keys[key]
	if !ok {
		return
	}
	for _, index := range itemIndices {
		delete(c.cells[index.cellLevel][index.cellPosition], key)
	}
	delete(c.keys, key)
}

// SearchCoveringResult are the boundaries of the cells used in the requested search
type SearchCoveringResult [][][]float64

// SearchCoveringParameters controls the algorithm and parameters used by S2 to determine the covering for the
// requested search area
type SearchCoveringParameters struct {
	LevelMod        int  `json:"level_mod"`
	MaxCells        int  `json:"max_cells"`
	MaxLevel        int  `json:"max_level"`
	MinLevel        int  `json:"min_level"`
	UseFastCovering bool `json:"use_fast_covering"`
}

// ItemsWithinDistance returns all contents stored in the collection within distanceMeters radius from the provided
// latitude an longitude. Note that this is an approximation and items further than distanceMeters may be returned, but
// it is guaranteed that all item ids returned are within distanceMeters. The caller of this function
// must specify all parameters used to generate cell covering as well as whether or not the coverer will use the
// standard covering algorithm or the fast covering algorithm which may be less precise.
func (c Collection) ItemsWithinDistance(
	latitude, longitude, distanceMeters float64, params SearchCoveringParameters,
) ([]interface{}, SearchCoveringResult) {
	// First, generate a spherical cap with an arc length of distanceMeters centered on the given latitude/longitude
	// This is the angle required (in radians) to trace an arc length of distanceMeters on the surface of the sphere
	capAngle := s1.Angle(distanceMeters / EarthRadiusMeters)
	capCenter := NewPointFromLatLng(latitude, longitude)
	searchCap := s2.CapFromCenterAngle(capCenter, capAngle)

	coverer := s2.RegionCoverer{
		MaxLevel: params.MaxLevel,
		MinLevel: params.MinLevel,
		LevelMod: params.LevelMod,
		MaxCells: params.MaxCells,
	}
	region := s2.Region(searchCap)
	var cellUnion s2.CellUnion
	if params.UseFastCovering {
		cellUnion = coverer.FastCovering(region)
	} else {
		cellUnion = coverer.Covering(region)
	}

	c.mutex.RLock()
	defer c.mutex.RUnlock()
	foundItems := make([]interface{}, 0)
	cellBounds := make(SearchCoveringResult, 0, len(cellUnion))
	for _, cell := range cellUnion {
		// get vertices in counter-clockwise order starting from the lower left
		vertices := make([][]float64, 5)
		for i := 0; i < 4; i++ {
			vertex := s2.CellFromCellID(cell).Vertex(i)
			ll := s2.LatLngFromPoint(vertex)
			vertices[i] = []float64{ll.Lng.Degrees(), ll.Lat.Degrees()}
		}
		// close the polygon loop
		vertices[4] = vertices[0]
		cellBounds = append(cellBounds, vertices)
		for key := range c.cells[cell.Level()][cell.Pos()] {
			foundItems = append(foundItems, c.items[key].contents)
		}
	}

	return foundItems, SearchCoveringResult(cellBounds)
}

// ItemByKey returns the contents stored in the collection by its key instead of by a geolocation lookup
func (c Collection) ItemByKey(key interface{}) interface{} {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	contents, ok := c.items[key]
	if !ok {
		return nil
	}
	return contents.contents
}

// NewPointFromLatLng constructs an s2 point from a lat/lon ordered pair
func NewPointFromLatLng(latitude, longitude float64) s2.Point {
	latLng := s2.LatLngFromDegrees(latitude, longitude)
	return s2.PointFromLatLng(latLng)
}

// EarthDistanceMeters calculates the distance in meters between two points on the surface of the Earth
func EarthDistanceMeters(p1, p2 s2.Point) float64 {
	return float64(p1.Distance(p2)) * EarthRadiusMeters
}
