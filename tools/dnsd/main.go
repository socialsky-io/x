package main

import (
	"bufio"
	"flag"
	"log"
	"net/http"

	"os"
	"os/signal"
	"syscall"

	"github.com/Xe/x/internal"
	"github.com/miekg/dns"
	"github.com/mmikulicic/stringlist"
)

var (
	port   = flag.String("port", "53", "UDP port to listen on for DNS")
	server = flag.String("forward-server", "1.1.1.1:53", "forward DNS server")

	zoneURLs = stringlist.Flag("zone-url", "DNS zonefiles to load")
)

var (
	defaultZoneURLS = []string{
		"https://xena.greedo.xeserv.us/files/akua.zone",
		"https://xena.greedo.xeserv.us/files/adblock.zone",
	}
)

func main() {
	internal.HandleStartup()

	if len(*zoneURLs) == 0 {
		*zoneURLs = defaultZoneURLS
	}

	for _, zurl := range *zoneURLs {
		log.Printf("conf: -zone-url=%s", zurl)
	}
	log.Printf("conf: -port=%s", *port)
	log.Printf("conf: -forward-server=%s", *server)

	rrs := []dns.RR{}

	for _, zurl := range *zoneURLs {
		resp, err := http.Get(zurl)
		if err != nil {
			panic(err)
		}

		reader := bufio.NewReaderSize(resp.Body, 2048)

		var i int
		zp := dns.NewZoneParser(reader, "", zurl)
		for rr, ok := zp.Next(); ok; rr, ok = zp.Next() {
			rrs = append(rrs, rr)
			i++
		}

		if zp.Err() != nil {
			panic(zp.Err())
		}

		resp.Body.Close()

		log.Printf("%s: %d records", zurl, i)
	}

	dns.HandleFunc(".", func(w dns.ResponseWriter, r *dns.Msg) {
		m := new(dns.Msg)
		m.SetReply(r)
		m.Authoritative = true

		for _, q := range r.Question {
			answers := []dns.RR{}
			for _, rr := range rrs {
				rh := rr.Header()

				if rh.Rrtype == dns.TypeCNAME && q.Name == rh.Name {
					answers = append(answers, rr)

					for _, a := range resolver("127.0.0.1:"+*port, rr.(*dns.CNAME).Target, q.Qtype) {
						answers = append(answers, a)
					}
				}

				if q.Name == rh.Name && q.Qtype == rh.Rrtype && q.Qclass == rh.Class {
					answers = append(answers, rr)
				}
			}
			if len(answers) == 0 && *server != "" {
				for _, a := range resolver(*server, q.Name, q.Qtype) {
					answers = append(answers, a)
				}
			}
			for _, a := range answers {
				m.Answer = append(m.Answer, a)
			}
		}
		w.WriteMsg(m)
	})

	go func() {
		srv := &dns.Server{Addr: ":" + *port, Net: "udp"}
		if err := srv.ListenAndServe(); err != nil {
			log.Fatalf("Failed to set udp listener %s\n", err.Error())
		}
	}()

	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	s := <-sig
	log.Fatalf("Signal (%v) received, stopping\n", s)
}

func resolver(server, fqdn string, r_type uint16) []dns.RR {
	m1 := new(dns.Msg)
	m1.Id = dns.Id()
	m1.SetQuestion(fqdn, r_type)

	in, err := dns.Exchange(m1, server)
	if err == nil {
		return in.Answer
	}
	return []dns.RR{}
}