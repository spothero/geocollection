// Copyright 2018 SpotHero
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
	"encoding/gob"

	"github.com/stretchr/testify/mock"
)

// MockGeoLocationCache mocks the GeoLocationCache implementation for use in tests
type MockGeoLocationCache struct {
	mock.Mock
}

func init() {
	gob.Register(&MockGeoLocationCache{})
}

// ItemsWithinDistance is a mocked version of ItemsWithinDistance
func (m *MockGeoLocationCache) ItemsWithinDistance(
	latitude, longitude, distanceMeters float64, params SearchCoveringParameters,
) ([]int, SearchCoveringResult) {
	args := m.Called(latitude, longitude, distanceMeters)
	return args.Get(0).([]int), args.Get(1).(SearchCoveringResult)
}

// Set is a mocked version of Set
func (m *MockGeoLocationCache) Set(id int, latitude, longitude float64) {
	m.Called(id, latitude, longitude)
}

// Delete is a mocked version of Delete
func (m *MockGeoLocationCache) Delete(id int) {
	m.Called(id)
}
