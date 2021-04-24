// Copyright (C) 2020-2021 digital-comrades and others
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

import "io"

// Connection interface represents a connection between two peers.
// The connection can be incoming, outgoing or duplex.
type Connection interface {
	io.Closer

	// Write sends the encoded date to the target peer.
	Write([]byte) error

	// Listen start listening for incoming data.
	Listen()
}
