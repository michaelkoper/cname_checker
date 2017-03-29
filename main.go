package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/miekg/dns"
)

var nusiiFQDN = []string{"nusii", "com"}

func main() {
	var inputs [][2]string
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		f := strings.Fields(scanner.Text())
		if len(f) == 0 {
			log.Fatalf("invalid input %q", scanner.Text())
		}

		var expect string
		if len(f) > 1 {
			expect = f[1]
		}
		inputs = append(inputs, [2]string{f[0], expect})
	}

	fmt.Printf("Checking %d hosts\n", len(inputs))

	var errors []string
	for _, input := range inputs {
		if err := validateHost(input[0], input[1]); err != nil {
			errors = append(errors, fmt.Sprintf("❗️\t%s\t%v", input[0], err))
		} else {
			fmt.Printf("✅\t%s\n", input[0])
		}
	}

	for _, err := range errors {
		fmt.Println(err)
	}
}

func validateHost(host, expected string) error {
	q := &dns.Msg{
		MsgHdr: dns.MsgHdr{
			Id:               dns.Id(),
			RecursionDesired: true,
		},
		Question: []dns.Question{
			{
				Name:   dns.Fqdn(host),
				Qtype:  dns.TypeCNAME,
				Qclass: dns.ClassINET,
			},
		},
	}

	in, err := dns.Exchange(q, "8.8.8.8:53")
	if err != nil {
		return err
	}

	var cnameFound bool
	for _, a := range in.Answer {
		if cname, ok := a.(*dns.CNAME); ok {
			cnameFound = true

			target := strings.TrimRight(cname.Target, ".")
			parts := strings.Split(target, ".")
			if expected != "" {
				if len(parts) == 0 || parts[0] != expected {
					return fmt.Errorf("got %q; expected %q", parts[0], expected)
				}
			}

			fqdn := parts[1:]
			if got, expected := len(fqdn), len(nusiiFQDN); got != expected {
				return fmt.Errorf("got %v; expected %v", fqdn, nusiiFQDN)
			}
			for i := range nusiiFQDN {
				if nusiiFQDN[i] != fqdn[i] {
					return fmt.Errorf("got %v; expected %v", fqdn, nusiiFQDN)
				}
			}
		}
	}

	if !cnameFound {
		return fmt.Errorf("no CNAME for %q", host)
	}

	return nil
}
