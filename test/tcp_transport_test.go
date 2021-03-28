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

package test

import (
	"context"
	"github.com/digital-comrades/proletariat/pkg/proletariat"
	"testing"
)

func TestTCPTransport_GoodAddress(t *testing.T) {
	_, err := proletariat.NewTCPTransport(context.TODO(), "localhost:0")
	if err != nil {
		t.Fatalf("failed: %v", err)
	}
}

func TestTCPTransport_BadAddress(t *testing.T) {
	_, err := proletariat.NewTCPTransport(context.TODO(), "0.0.0.0:0")
	if err != proletariat.ErrInvalidAddr {
		t.Fatalf("failed: %v", err)
	}
}

func TestTCPTransport_EmptyAddr(t *testing.T) {
	_, err := proletariat.NewTCPTransport(context.TODO(), ":0")
	if err != proletariat.ErrInvalidAddr {
		t.Fatalf("failed: %v", err)
	}
}
