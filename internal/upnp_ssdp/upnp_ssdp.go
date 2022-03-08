// package upnp_ssdp allows Workers to find their Manager on the LAN.
package upnp_ssdp

/* ***** BEGIN GPL LICENSE BLOCK *****
 *
 * Original Code Copyright (C) 2022 Blender Foundation.
 *
 * This file is part of Flamenco.
 *
 * Flamenco is free software: you can redistribute it and/or modify it under
 * the terms of the GNU General Public License as published by the Free Software
 * Foundation, either version 3 of the License, or (at your option) any later
 * version.
 *
 * Flamenco is distributed in the hope that it will be useful, but WITHOUT ANY
 * WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR
 * A PARTICULAR PURPOSE.  See the GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License along with
 * Flamenco.  If not, see <https://www.gnu.org/licenses/>.
 *
 * ***** END GPL LICENSE BLOCK ***** */

const (
	FlamencoUUID        = "aa80bc5f-d0af-46b8-8630-23bd7e80ec4d"
	FlamencoServiceType = "urn:flamenco:manager:0"

	serviceDescriptionPath = "/upnp/description.xml"
)

// SSDP Service description, usually served on some URL ending in `/description.xml`.
type Description struct {
	XMLName     string      `xml:"urn:schemas-upnp-org:device-1-0 root"`
	SpecVersion SpecVersion `xml:"specVersion"`
	URLBase     string      `xml:"URLBase"`
	Device      Device      `xml:"device"`
}
type SpecVersion struct {
	Major int `xml:"major"`
	Minor int `xml:"minor"`
}
type Device struct {
	DeviceType       string   `xml:"deviceType"`
	FriendlyName     string   `xml:"friendlyName"`
	Manufacturer     string   `xml:"manufacturer"`
	ManufacturerURL  string   `xml:"manufacturerURL"`
	ModelDescription string   `xml:"modelDescription"`
	ModelName        string   `xml:"modelName"`
	ModelURL         string   `xml:"modelURL"`
	UDN              string   `xml:"UDN"`
	ServiceList      []string `xml:"serviceList"` // not []string, but since the list is empty, it doesn't matter.
	PresentationURL  string   `xml:"presentationURL"`
}
