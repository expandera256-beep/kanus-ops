package main

// runAppWindow hosts the panel in the program's own window. It reports
// false when that window cannot be created, so the browser can open instead.
var runAppWindow = func(string) bool { return false }

// spawnAppWindow starts that window in another process. The listener stays
// up when the window cannot be created.
var spawnAppWindow = func(string) {}
