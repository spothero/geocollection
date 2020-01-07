// Copyright 2020 SpotHero
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

package testhelpers

import (
	"github.com/spothero/geocollection"
	"github.com/stretchr/testify/mock"
)

// MockCollection provides a mock LocationCollection for use in tests
type MockCollection struct {
	mock.Mock
}

// ItemsWithinDistance is a mocked version of ItemsWithinDistance
func (m *MockCollection) ItemsWithinDistance(latitude, longitude, distanceMeters float64, params geocollection.SearchCoveringParameters) ([]interface{}, geocollection.SearchCoveringResult) {
	args := m.Called(latitude, longitude, distanceMeters, params)
	return args.Get(0).([]interface{}), args.Get(1).(geocollection.SearchCoveringResult)
}

// Set is a mocked version of Set
func (m *MockCollection) Set(key, contents interface{}, latitude, longitude float64) {
	m.Called(key, contents, latitude, longitude)
}

// Delete is a mocked version of Delete
func (m *MockCollection) Delete(key interface{}) {
	m.Called(key)
}

func (m *MockCollection) ItemByKey(key interface{}) interface{} {
	return m.Called(key).Get(0)
}
