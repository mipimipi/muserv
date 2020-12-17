// SPDX-FileCopyrightText: 2018-2020 Michael Picht <mipi@fsfe.org>
//
// SPDX-License-Identifier: GPL-3.0-or-later

package utils

import "encoding/xml"

// MarshalXML wraps xml.Marshal() adding the prefix <?xml version="1.0"?>
func MarshalXML(a interface{}) (x []byte, err error) {
	xmlPrefix := []byte("<?xml version=\"1.0\"?>")

	if x, err = xml.Marshal(a); err != nil {
		return
	}
	return append(xmlPrefix, x...), nil
}

// MarshalIndentXML wraps xml.MarshalIndent() adding the prefix <?xml version="1.0"?>
func MarshalIndentXML(a interface{}, prefix string, indent string) (x []byte, err error) {
	xmlPrefix := []byte(xml.Header)

	if x, err = xml.MarshalIndent(a, prefix, indent); err != nil {
		return
	}
	return append(xmlPrefix, x...), nil
}
