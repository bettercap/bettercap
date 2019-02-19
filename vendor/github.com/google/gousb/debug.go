// Copyright 2013 Google Inc.  All rights reserved.
// Copyright 2016 the gousb Authors.  All rights reserved.
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

package gousb

// To enable internal debugging, set the GOUSB_DEBUG environment variable.

import (
	"io"
	"io/ioutil"
	"log" // TODO(kevlar): make a logger
	"os"
)

var debug *log.Logger

const debugEnvVarName = "GOUSB_DEBUG"

func init() {
	out := io.Writer(ioutil.Discard)
	if os.Getenv(debugEnvVarName) != "" {
		out = os.Stderr
	}
	debug = log.New(out, "gousb: ", log.LstdFlags|log.Lshortfile)
}
