// Command confighub starts a ConfigHub cluster process with an embedded console.
package main

import (
	"flag"
	"fmt"
	"log"

	"confighub/internal/cluster"
	"confighub/internal/settings"
)

func main() {
	var (
		addr    = flag.String("addr", "", "HTTP listen address")
		data    = flag.String("data", "", "data directory")
		quota   = flag.Int64("quota", 0, "default namespace quota in bytes")
		nsCount = flag.Int("namespaces", 0, "number of namespaces")
	)
	flag.Parse()
	cfg := settings.Default()
	if *addr != "" {
		cfg.HTTPAddr = *addr
	}
	if *data != "" {
		cfg.DataDir = *data
	}
	if *quota > 0 {
		cfg.DefaultQuota = *quota
	}
	if *nsCount > 0 {
		cfg.NamespaceCount = *nsCount
	}
	cl, err := cluster.Build(cfg)
	if err != nil {
		log.Fatalf("confighub: build: %v", err)
	}
	defer func() {
		if err := cl.Close(); err != nil {
			log.Printf("confighub: close: %v", err)
		}
	}()
	fmt.Printf("confighub listening on %s (data %s)\n", cfg.HTTPAddr, cfg.DataDir)
	if err := cl.Start(); err != nil {
		log.Fatal(err)
	}
}
