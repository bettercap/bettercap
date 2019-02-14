/*
   Copyright (c) 2012 Kyle Isom <kyle@tyrfingr.is>

   Permission to use, copy, modify, and distribute this software for any
   purpose with or without fee is hereby granted, provided that the
   above copyright notice and this permission notice appear in all
   copies.

   THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL
   WARRANTIES WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED
   WARRANTIES OF MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE
   AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL
   DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM LOSS OF USE, DATA
   OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR OTHER
   TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
   PERFORMANCE OF THIS SOFTWARE.
*/

/*
Package fswatch provides simple UNIX file system watching in Go. It is based around
the Watcher struct, which should be initialised with either NewWatcher or
NewAutoWatcher. Both functions accept a variable number of string arguments
specfying the paths to be loaded, which may be globbed, and return a pointer
to a Watcher. This value can be started and stopped with the Start() and
Stop() methods. The Watcher will automatically stop if all the files it is
watching have been deleted.

The Start() method returns a read-only channel that receives Notification
values. The Stop() method closes the channel, and no files will be watched
from that point.

The list of files being watched may be retrieved with the Watch() method and
the current state of the files being watched may be retrieved with the
State() method. See the go docs for more information.

In synchronous mode (i.e. Watchers obtained from NewWatcher()), deleted files
will not be removed from the watch list, allowing the user to watch for files
that might be created at a future time, or to allow notification of files that
are deleted and then recreated. The auto-watching mode (i.e. from
NewAutoWatcher()) will remove deleted files from the watch list, as it
automatically adds new files to the watch list.

If "." is not specified explicitly in the list of files to watch, new
directories created in the current directory will not be seen (as per the
behaviour of filepath.Match); any directories being watched will, however.
If you wish to watch for changes in the current directory, be sure to specify
".".
*/
package fswatch
