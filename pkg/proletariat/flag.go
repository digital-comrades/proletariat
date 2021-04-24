// Copyright (C) 2020-2021 digital-comrades and others.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//         http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package proletariat

import "sync/atomic"

const (
	active   = 0x0
	inactive = 0x1
)

// The Flag structure.
// Only some transitions are available when using this structure,
// so this will not act as an atomic boolean. The accepted transitions are:
//
// IsActive if flag is 0x0, returns `true`;
// IsInactive if flag is 0x1, returns `true`;
// Inactivate iff IsActive change value to IsInactive.
//
// The start value will be `active`.
type Flag struct {
	// Holds the current state of the flag.
	flag int32
}

// IsActive returns `true` if the flag still active.
func (f *Flag) IsActive() bool {
	return atomic.LoadInt32(&f.flag) == active
}

// IsInactive returns `true` if the flag is inactive.
func (f *Flag) IsInactive() bool {
	return atomic.LoadInt32(&f.flag) == inactive
}

// Inactivate will inactivate the flag.
// Returns `true` if was active and now is inactive and returns `false`
// if it was already inactivated.
func (f *Flag) Inactivate() bool {
	return atomic.CompareAndSwapInt32(&f.flag, active, inactive)
}
