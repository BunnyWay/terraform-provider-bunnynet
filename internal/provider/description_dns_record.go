// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package provider

type dnsRecordDescriptionType struct {
	Id                  string
	Zone                string
	Enabled             string
	Type                string
	TTL                 string
	Value               string
	Name                string
	Weight              string
	Priority            string
	Port                string
	Flags               string
	Tag                 string
	Accelerated         string
	AcceleratedPullzone string
	Link                string
	MonitorType         string
	GeolocationLat      string
	GeolocationLong     string
	LatencyZone         string
	SmartRoutingType    string
	Comment             string
}

var dnsRecordDescription = dnsRecordDescriptionType{
	Id:                  "The unique identifier for the DNS record.",
	Zone:                "ID of the related DNS zone.",
	Enabled:             "Indicates whether the DNS record is enabled.",
	Type:                generateMarkdownMapOptions(dnsRecordTypeMap),
	TTL:                 "The time-to-live value for the DNS record.",
	Value:               "The value of the DNS record.",
	Name:                `The name of the DNS record. Use <code>name = ""</code> for apex domain records.`,
	Weight:              "The weight of the DNS record. It is used in load balancing scenarios to distribute traffic based on the specified weight.",
	Priority:            "The priority of the DNS record.",
	Port:                "The port number for services that require a specific port.",
	Flags:               "Flags for advanced DNS settings.",
	Tag:                 "A tag for the DNS record.",
	Accelerated:         "Indicates whether the DNS record should utilize bunny.net’s acceleration services.",
	AcceleratedPullzone: "The ID of the accelerated pull zone.",
	Link:                "The name of the linked resource.",
	MonitorType:         generateMarkdownMapOptions(dnsRecordMonitorTypeMap),
	GeolocationLat:      "The latitude for geolocation-based routing.",
	GeolocationLong:     "The longitude for geolocation-based routing.",
	LatencyZone:         "The latency zone for latency-based routing.",
	SmartRoutingType:    generateMarkdownMapOptions(dnsRecordSmartRoutingTypeMap),
	Comment:             "This property allows users to add descriptive notes for documentation and management purposes.",
}
