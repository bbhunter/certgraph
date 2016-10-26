package main

import (
	"bufio"
	"fmt"
	"gopkg.in/jmcvetta/neoism.v1"
	"log"
	"os"
	"strings"
)

var db *neoism.Database

/*
idea
pass 1 insert domains, certs, and domain-> cert
pass 2 insert certs-> domains
*/

func main() {
	// Connect to the Neo4j server
	var err error
	db, err = neoism.Connect("http://neo4j:PASSWORD@localhost:7474/db/data")
	check(err)

	if len(os.Args) != 2 {
		log.Fatal("Pass file to parse")
	}

	// pass 1
	fmt.Println("pass1")
	pass1(os.Args[1])

	// pass 2
	fmt.Println("pass2")
	pass2(os.Args[1])

	fmt.Println("done")
}

func tabSplit(in string) []string {
	out := make([]string, strings.Count(in, "\t")+1)
	n := 0
	for i := range out {
		p := strings.Index(in[n:], "\t")
		if p == -1 {
			p = len(in)
		} else {
			p = p + n
		}
		if p > n {
			out[i] = in[n:p]
		}
		n = p + 1
	}
	return out
}

func pass1(path string) {
	file, err := os.Open(path)
	check(err)
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := scanner.Text()
		parts := tabSplit(line)
		fmt.Println(parts[0])
		addDomainCert(parts[0], parts[2], parts[3])
	}
}

func pass2(path string) {
	file, err := os.Open(path)
	check(err)
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := scanner.Text()
		parts := tabSplit(line)
		domains := parseDomainArray(parts[4])
		fmt.Println("pass2", parts[0], len(domains))
		addCertDomains(parts[3], domains)
	}
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func parseDomainArray(str string) []string {
	if str[0] != '[' {
		log.Fatal("Domain array does not start with [")
	}
	if str[len(str)-1] != ']' {
		log.Fatal("Domain array does not end with ]")
	}
	str = str[1 : len(str)-1]
	return strings.Split(str, " ")
}

func addDomainCert(domain, status, cert string) {
	// add domain if new
	// TODO is label created?
	domainNode := GetOrCreateNode("domain", "domain", neoism.Props{"domain": directDomain(domain), "status": status})
	//check(err)

	if len(cert) > 0 {

		// add cert if new
		certNode := GetOrCreateNode("certificate", "fingerprint", neoism.Props{"fingerprint": cert})
		//check(err)

		// add domain -> cert
		GetOrCreateRelationship(domainNode, certNode, "uses", neoism.Props{})
		//_, err = domainNode.Relate("has cert", certNode.Id(), neoism.Props{})
		//check(err)
	}
}

func addCertDomains(cert string, domains []string) {
	// add cert if new
	certNode := GetOrCreateNode("certificate", "fingerprint", neoism.Props{"fingerprint": cert})
	//check(err)

	//fmt.Println(domains)
	for _, domain := range domains {
		domainNode := GetOrCreateNode("domain", "domain", neoism.Props{"domain": directDomain(domain)})
		//check(err)

		/*if created {
			log.Fatal("inserted domain on 2nd pass", domain)
		}*/

		// TODO error check?
		GetOrCreateRelationship(certNode, domainNode, "SAN", neoism.Props{"domain": domain})
	}
}

// given a domain returns the non-wildcard version of that domain
func directDomain(domain string) string {
	if len(domain) < 3 {
		return domain
	}
	if domain[0:2] == "*." {
		domain = domain[2:]
	}
	return domain
}

//from: https://github.com/CSharpRU/neo4j-last-fm-importer/blob/master/src/importer/neo4j.go
// TODO refractor
func GetOrCreateRelationship(from *neoism.Node, to *neoism.Node, relType string, props neoism.Props) (relationship *neoism.Relationship) {
	relationships, err := from.Relationships(relType)

	if err == nil {
		for _, relationship := range relationships {
			endNode, err := relationship.End()

			if err != nil {
				continue
			}

			if endNode.Id() == to.Id() {
				/*newProps, err := relationship.Properties()

				if err != nil {
					return relationship
				}*/

				/*if err := mergo.Merge(&newProps, props); err != nil {
					relationship.SetProperties(newProps)
				}*/

				return relationship
			}
		}
	}

	relationship, err = from.Relate(relType, to.Id(), props)
	check(err)

	return
}

func GetOrCreateNode(label, key string, props neoism.Props) (node *neoism.Node) {
	node, created, err := db.GetOrCreateNode(label, key, props)

	check(err)

	if created {
		err := node.AddLabel(label)

		check(err)
	}

	return
}
