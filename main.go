package main

import (
    "flag"
    "fmt"
    "github.io/vsverchkov/d-cache/cache"
    "github.io/vsverchkov/d-cache/client"
    "log"
    "time"
)

func main() {
    var (
        listenAddr = flag.String("listenaddr", ":3000", "listen address of the server")
        leaderAddr = flag.String("leaderaddr", "", "listen address of the leader")
    )
    flag.Parse()

    opts := ServerOpts{
        ListenAddr: *listenAddr,
        IsLeader: len(*leaderAddr) == 0,
        LeaderAddr: *leaderAddr,
    }

    go func() {
        time.Sleep(time.Second * 10)
        if opts.IsLeader {
            PerformCache()
        }
    }()

	server := NewServer(opts, cache.New())
    server.Start()
}

func PerformCache() {
    for i := 0; i < 100; i++ {
        go func(i int) {
            newClient, err := client.New(":3000")
            if err != nil {
                log.Fatal(err)
            }

            var (
                key   = []byte(fmt.Sprintf("key_%d", i))
                value = []byte(fmt.Sprintf("val_%d", i))
            )

            err = newClient.Set(key, value, 10000000000)
            if err != nil {
                log.Fatal(err)
            }

            fetchedValue, err := newClient.Get(key)
            if err != nil {
                log.Fatal(err)
            }
            fmt.Println(string(fetchedValue))

            newClient.Close()
        }(i)
    }
}