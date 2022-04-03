package coredns_gdns

import "fmt"

type IpLocation struct {
	location string
}

func NewIpLocation() *IpLocation {
	return nil
}

func (iplocation *IpLocation) parse(ip string) (string, error) {
	fmt.Println(iplocation)
	return "", nil
}
