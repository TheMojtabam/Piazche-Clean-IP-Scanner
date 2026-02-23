package utils

import (
	"bufio"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strings"
)

// ParseIPsFromFile reads IP addresses and CIDR blocks from a file
func ParseIPsFromFile(path string, sampleSize int, shuffle bool) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var ips []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.Contains(line, "/") {
			subnetIPs, err := ExpandCIDR(line, sampleSize)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: invalid CIDR %s: %v\n", line, err)
				continue
			}
			ips = append(ips, subnetIPs...)
		} else {
			if net.ParseIP(line) != nil {
				ips = append(ips, line)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	if shuffle {
		ShuffleIPs(ips)
	}

	return ips, nil
}

// ExpandCIDR expands a CIDR block to individual IP addresses.
// Optionally samples random IPs if sampleSize > 0.
func ExpandCIDR(cidr string, sampleSize int) ([]string, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	var ips []string
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); incrementIP(ip) {
		// Skip network and broadcast addresses for typical subnets
		ones, bits := ipnet.Mask.Size()
		if ones >= 24 && bits == 32 {
			lastOctet := ip[3]
			if lastOctet == 0 || lastOctet == 255 {
				continue
			}
		}
		ips = append(ips, ip.String())
	}

	if sampleSize > 0 && sampleSize < len(ips) {
		return SampleIPs(ips, sampleSize), nil
	}

	return ips, nil
}

// incrementIP increments an IP address by 1
func incrementIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

// SampleIPs randomly samples n IPs from the list
func SampleIPs(ips []string, n int) []string {
	if n >= len(ips) {
		return ips
	}

	result := make([]string, len(ips))
	copy(result, ips)

	for i := 0; i < n; i++ {
		j := i + rand.Intn(len(result)-i)
		result[i], result[j] = result[j], result[i]
	}

	return result[:n]
}

// ShuffleIPs shuffles the IP list in place
func ShuffleIPs(ips []string) {
	rand.Shuffle(len(ips), func(i, j int) {
		ips[i], ips[j] = ips[j], ips[i]
	})
}

// ParseCIDRList parses a comma-separated list of CIDRs
func ParseCIDRList(input string, sampleSize int) ([]string, error) {
	var ips []string
	cidrs := strings.Split(input, ",")

	for _, cidr := range cidrs {
		cidr = strings.TrimSpace(cidr)
		if cidr == "" {
			continue
		}

		if strings.Contains(cidr, "/") {
			subnetIPs, err := ExpandCIDR(cidr, sampleSize)
			if err != nil {
				return nil, fmt.Errorf("invalid CIDR %s: %w", cidr, err)
			}
			ips = append(ips, subnetIPs...)
		} else {
			if net.ParseIP(cidr) != nil {
				ips = append(ips, cidr)
			} else {
				return nil, fmt.Errorf("invalid IP: %s", cidr)
			}
		}
	}

	return ips, nil
}

// IsValidIP checks if a string is a valid IP address
func IsValidIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

// IsValidCIDR checks if a string is a valid CIDR block
func IsValidCIDR(cidr string) bool {
	_, _, err := net.ParseCIDR(cidr)
	return err == nil
}
