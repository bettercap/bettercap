// Copyright 2013 Google Inc.  All rights reserved.
// Copyright 2018 the gousb Authors.  All rights reserved.
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

#include <libusb.h>

void gousb_set_debug(libusb_context *ctx, int lvl) {
    // TODO(sebek): remove libusb_debug entirely in 2.1 or 3.0,
    // require libusb >= 1.0.22. libusb 1.0.22 sets API version 0x01000106.
#if LIBUSB_API_VERSION >= 0x01000106
    libusb_set_option(ctx, LIBUSB_OPTION_LOG_LEVEL, lvl);
#else
    libusb_set_debug(ctx, lvl);
#endif
}
