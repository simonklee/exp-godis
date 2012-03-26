package main

import (
    "time"
    "fmt"
    "os"
    "github.com/simonz05/redis"
    "strings"
    "flag"
    "runtime"
    "runtime/pprof"
)

var tests = make(map[string]func(*redis.Client, chan bool))
var C *int = flag.Int("c", 50, "concurrent requests")
var R *int = flag.Int("r", 4, "sample size")
var N *int = flag.Int("n", 10000, "number of request")

func init() {
    runtime.GOMAXPROCS(8)

    tests["set"] = setHandle
    tests["get"] = getHandle
}

func prints(t time.Duration) {
    fmt.Fprintf(os.Stdout, "    %.2f op/sec  real %.4fs\n", float64(*N)/t.Seconds(), t.Seconds())
}

func printsA(avg, tot time.Duration) {
    fmt.Fprintf(os.Stdout, "%.2f op/sec  real %.4fs  tot %.4fs\n", float64(*N)/avg.Seconds(), avg.Seconds(), tot.Seconds())
}

func setHandle(c *redis.Client, ch chan bool) {
    for _ = range ch {
        c.Call("SET", "foo", "bar")
    }
}

func getHandle(c *redis.Client, ch chan bool) {
    for _ = range ch {
        c.Call("GET", "foo")
    }
}

func BenchmarkRedis(handle func(*redis.Client, chan bool)) time.Duration {
    c := redis.NewClient("")
    ch := make(chan bool)
    start := time.Now()

    for i := 0; i < *C; i++ {
        go handle(c, ch)
    }

    for i := 0; i < *N; i++ {
        ch<-true 
    }

    return time.Now().Sub(start)
}

func run(name string) {
    var t, total time.Duration
    test, ok := tests[name]

    if !ok {
        fmt.Fprintf(os.Stderr, "test: `%s` does not exists\n", name)
        os.Exit(1)
    }

    fmt.Printf("%s:\n",strings.ToUpper(name))

    for i := 0; i < *R; i++ {
        t = BenchmarkRedis(test)
        total += t
        prints(t)
    }

    avg := time.Duration(total.Nanoseconds() / int64(*R))

    print("AVG ")
    printsA(avg, total)
    println()
}

func main() {
    flag.Parse()
    fmt.Printf("CONCURRENT: %d SAMPLES: %d REQUESTS: %d\n\n", *C, *R, *N)

    for _, name := range flag.Args() {
        run(name)
    }

    stats := new(runtime.MemStats)
    runtime.ReadMemStats(stats)
    pprof.StopCPUProfile()
}