# fswatch
## go library for simple UNIX file system watching

fswatch provides simple UNIX file system watching in Go. It is based around
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

## Usage
There are two types of Watchers:

* static watchers watch a limited set of files; they do not purge deleted
files from the watch list.
* auto watchers watch a set of files and directories; directories are
watched for new files. New files are automatically added, and deleted
files are removed from the watch list.

Take a look at the provided `clinotify/clinotify.go` for an example; the
package is also well-documented. See the godocs for more specifics.

## License

`fswatch` is licensed under the ISC license.
