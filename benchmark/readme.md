# FJSON benchmark

First, build the project: `go build -ldflags "-w -s"`.

**Command line options:**
- `host`: the FJSON server host (default `localhost`)
- `port`: the FJSON server port (default `8080`)
- `routines`: the number of goroutine used for the benchmark (default `1`). It is advised to not use more routines than your CPU has available threads.
- `connections`: the total number of connections to keep with each routing handling N = connections/routines open (default `400`).
- `duration`: the number of seconds the benchmark will run for (default `30`)

**Example:**
```
./benchmark -host=127.0.0.1 -port=8080 -routines=64 -duration=30
```