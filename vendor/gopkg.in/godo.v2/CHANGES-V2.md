v2.0.4 / 2016-01-14
===================

  * remove commented code
  * Context.Start: improve rebuild time on watch by building changed file's package only instead of using -a flag

v2.0.3 / 2015-12-10
===================

  * update README
  * fix godoenv parsing on rebuild
[x] Tasks have Src -> Dest to more efficiently watch and rebuild

[x] Run dependencies in Parallel or Series

[x] Godo will search up dir tree for nearest Gododir/main.go

[x] Namespaces to better manage or import tasks

[x] Optimize watch algorithm

[x] Allow exec commands to be teed, print or captured

[x]  More efficient file watcher

[x]  Externalize glob

[x]  Deprecated

    In{},
    D{}
    W{}
    c.Args.ZeroString -> c.Args.AsString


[x] Set environment variables via key=value pairs

[x] Watches Godofile (Gododir/main.go) automatically (buggy)

