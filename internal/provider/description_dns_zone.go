// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package provider

type dnsZoneDescriptionType struct {
	Id                 string
	Domain             string
	NameserverCustom   string
	Nameserver1        string
	Nameserver2        string
	SoaEmail           string
	LogEnabled         string
	LogAnonymized      string
	LogAnonymizedStyle string
}

var dnsZoneDescription = dnsZoneDescriptionType{
	Id:                 "The unique identifier for the DNS zone.",
	Domain:             "The domain name for the DNS zone.",
	NameserverCustom:   "Indicates whether custom nameservers are used.",
	Nameserver1:        "The primary nameserver for the DNS zone.",
	Nameserver2:        "The secondary nameserver for the DNS zone.",
	SoaEmail:           "The email address used in the Start of Authority (SOA) record for the DNS zone.",
	LogEnabled:         "Indicates whether permanent logging for DNS queries is enabled.",
	LogAnonymized:      "Indicates whether DNS logs are anonymized.",
	LogAnonymizedStyle: generateMarkdownMapOptions(pullzoneLogAnonymizedStyleMap),
}
