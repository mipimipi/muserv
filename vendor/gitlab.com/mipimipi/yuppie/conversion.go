// SPDX-FileCopyrightText: 2013 John Beisley <johnbeisleyuk@gmail.com>,
//
// SPDX-License-Identifier: BSD-2-Clause
//
// copied from https://github.com/huin/goupnp
// modified by Michael Picht <mipi@fsfe.org>

package yuppie

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/pkg/errors"
)

var (
	// localLoc acts like time.Local for this package, but is faked out by the
	// unit tests to ensure that things stay constant (especially when running
	// this test in a place where local time is UTC which might mask bugs).
	localLoc = time.Local
)

// marshalUpnpUI1 marshals ui1
func marshalUpnpUI1(v uint8) (string, error) {
	return strconv.FormatUint(uint64(v), 10), nil
}

// unmarshalUpnpUI1 unmarshals ui1
func unmarshalUpnpUI1(s string) (uint8, error) {
	if len(s) == 0 {
		return 0, nil
	}
	v, err := strconv.ParseUint(s, 10, 8)
	if err != nil {
		return 0, errors.Wrapf(err, "cannot unmarshal UI1 from '%s'", s)
	}
	return uint8(v), nil
}

// marshalUpnpUI2 marshals ui2
func marshalUpnpUI2(v uint16) (string, error) {
	return strconv.FormatUint(uint64(v), 10), nil
}

// unmarshalUpnpUI2 unmarshals ui2
func unmarshalUpnpUI2(s string) (uint16, error) {
	if len(s) == 0 {
		return 0, nil
	}
	v, err := strconv.ParseUint(s, 10, 16)
	if err != nil {
		return 0, errors.Wrapf(err, "cannot unmarshal UI2 from '%s'", s)
	}
	return uint16(v), nil
}

// marshalUpnpUI4 marshals ui4
func marshalUpnpUI4(v uint32) (string, error) {
	return strconv.FormatUint(uint64(v), 10), nil
}

// unmarshalUpnpUI4 unmarshals ui4
func unmarshalUpnpUI4(s string) (uint32, error) {
	if len(s) == 0 {
		return 0, nil
	}
	v, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0, errors.Wrapf(err, "cannot unmarshal UI4 from '%s'", s)
	}
	return uint32(v), nil
}

// marshalUpnpUI8 marshals ui8
func marshalUpnpUI8(v uint64) (string, error) {
	return strconv.FormatUint(v, 10), nil
}

// unmarshalUpnpUI8 unmarshals ui8
func unmarshalUpnpUI8(s string) (uint64, error) {
	if len(s) == 0 {
		return 0, nil
	}
	v, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, errors.Wrapf(err, "cannot unmarshal UI8 from '%s'", s)
	}
	return v, nil
}

// marshalUpnpI1 marshals i1
func marshalUpnpI1(v int8) (string, error) {
	return strconv.FormatInt(int64(v), 10), nil
}

// unmarshalUpnpI1 unmarshals i1
func unmarshalUpnpI1(s string) (int8, error) {
	if len(s) == 0 {
		return 0, nil
	}
	v, err := strconv.ParseInt(s, 10, 8)
	if err != nil {
		return 0, errors.Wrapf(err, "cannot unmarshal I1 from '%s'", s)
	}
	return int8(v), nil
}

// marshalUpnpI2 marshals i2
func marshalUpnpI2(v int16) (string, error) {
	return strconv.FormatInt(int64(v), 10), nil
}

// unmarshalUpnpI2 unmarshals i2
func unmarshalUpnpI2(s string) (int16, error) {
	if len(s) == 0 {
		return 0, nil
	}
	v, err := strconv.ParseInt(s, 10, 16)
	if err != nil {
		return 0, errors.Wrapf(err, "cannot unmarshal I2 from '%s'", s)
	}
	return int16(v), nil
}

// marshalUpnpI4 marshals i4
func marshalUpnpI4(v int32) (string, error) {
	return strconv.FormatInt(int64(v), 10), nil
}

// unmarshalUpnpI4 unmarshals i4
func unmarshalUpnpI4(s string) (int32, error) {
	if len(s) == 0 {
		return 0, nil
	}
	v, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return 0, errors.Wrapf(err, "cannot unmarshal I4 from '%s'", s)
	}
	return int32(v), nil
}

// marshalUpnpInt marshals int
func marshalUpnpInt(v int64) (string, error) {
	return strconv.FormatInt(v, 10), nil
}

// unmarshalUpnpInt unmarshals int
func unmarshalUpnpInt(s string) (int64, error) {
	if len(s) == 0 {
		return 0, nil
	}
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, errors.Wrapf(err, "cannot unmarshal Int from '%s'", s)
	}
	return v, nil
}

// marshalUpnpR4 marshals r4
func marshalUpnpR4(v float32) (string, error) {
	return strconv.FormatFloat(float64(v), 'G', -1, 32), nil
}

// unmarshalUpnpR4 unmarshals R4
func unmarshalUpnpR4(s string) (float32, error) {
	if len(s) == 0 {
		return 0, nil
	}
	v, err := strconv.ParseFloat(s, 32)
	if err != nil {
		return 0, errors.Wrapf(err, "cannot unmarshal R4 from '%s'", s)
	}
	return float32(v), nil
}

// marshalUpnpR8 marshals r8
func marshalUpnpR8(v float64) (string, error) {
	return strconv.FormatFloat(v, 'G', -1, 64), nil
}

// unmarshalUpnpR8 unmarshals R8
func unmarshalUpnpR8(s string) (float64, error) {
	if len(s) == 0 {
		return 0, nil
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, errors.Wrapf(err, "cannot unmarshal R8 from '%s'", s)
	}
	return v, nil
}

// marshalUpnpNumber marshals number
func marshalUpnpNumber(v float64) (string, error) {
	return strconv.FormatFloat(v, 'G', -1, 64), nil
}

// unmarshalUpnpNumber unmarshals Number
func unmarshalUpnpNumber(s string) (float64, error) {
	if len(s) == 0 {
		return 0, nil
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, errors.Wrapf(err, "cannot unmarshal Number from '%s'", s)
	}
	return v, nil
}

// marshalUpnpFixed14_4 marshals float64 to SOAP "fixed.14.4" type.
func marshalUpnpFixed14_4(v float64) (string, error) {
	if v >= 1e14 || v <= -1e14 {
		return "", fmt.Errorf("cannot marshal fixed14.4: value %v out of bounds", v)
	}
	return strconv.FormatFloat(v, 'f', 4, 64), nil
}

// unmarshalUpnpFixed14_4 unmarshals float64 from SOAP "fixed.14.4" type.
func unmarshalUpnpFixed14_4(s string) (float64, error) {
	if len(s) == 0 {
		return 0, nil
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, errors.Wrapf(err, "cannot unmarshal fixed.14.4 from '%s'", s)
	}
	if v >= 1e14 || v <= -1e14 {
		return 0, fmt.Errorf("cannot unmarshal fixed14.4: value %s out of bounds", s)
	}
	return v, nil
}

// marshalUpnpFloat marshals float64 to SOAP "fixed.14.4" type.
func marshalUpnpFloat(v float64) (string, error) {
	return marshalUpnpFixed14_4(v)
}

// unmarshalFloat unmarshals float64 from SOAP "fixed.14.4" type.
func unmarshalUpnpFloat(s string) (float64, error) {
	if len(s) == 0 {
		return 0, nil
	}
	return unmarshalUpnpFixed14_4(s)
}

// marshalUpnpChar marshals rune to SOAP "char" type.
func marshalUpnpChar(v rune) (string, error) {
	if v == 0 {
		return "", fmt.Errorf("cannot marshal Char: rune 0 is not allowed")
	}
	return string(v), nil
}

// unmarshalUpnpChar unmarshals rune from SOAP "char" type.
func unmarshalUpnpChar(s string) (rune, error) {
	if len(s) == 0 {
		return 0, fmt.Errorf("cannot unmarshal Char: got empty string")
	}
	r, n := utf8.DecodeRune([]byte(s))
	if n != len(s) {
		return 0, fmt.Errorf("cannot unmarshal Char: value %s is not a single rune", s)
	}
	return r, nil
}

// marshalUpnpString marshals string
func marshalUpnpString(v string) (string, error) {
	return v, nil
}

// unmarshalUpnpString unmarshals string
func unmarshalUpnpString(v string) (string, error) {
	return v, nil
}

func parseInt(s string, err *error) int {
	v, parseErr := strconv.ParseInt(s, 10, 64)
	if parseErr != nil {
		*err = parseErr
	}
	return int(v)
}

var dateRegexps = []*regexp.Regexp{
	// yyyy[-mm[-dd]]
	regexp.MustCompile(`^(\d{4})(?:-(\d{2})(?:-(\d{2}))?)?$`),
	// yyyy[mm[dd]]
	regexp.MustCompile(`^(\d{4})(?:(\d{2})(?:(\d{2}))?)?$`),
}

func parseDateParts(s string) (year, month, day int, err error) {
	var parts []string
	for _, re := range dateRegexps {
		parts = re.FindStringSubmatch(s)
		if parts != nil {
			break
		}
	}
	if parts == nil {
		err = fmt.Errorf("soap date: value %q is not in a recognized ISO8601 date format", s)
		return
	}

	year = parseInt(parts[1], &err)
	month = 1
	day = 1
	if len(parts[2]) != 0 {
		month = parseInt(parts[2], &err)
		if len(parts[3]) != 0 {
			day = parseInt(parts[3], &err)
		}
	}

	if err != nil {
		err = fmt.Errorf("soap date: %q: %v", s, err)
	}

	return
}

var timeRegexps = []*regexp.Regexp{
	// hh[:mm[:ss]]
	regexp.MustCompile(`^(\d{2})(?::(\d{2})(?::(\d{2}))?)?$`),
	// hh[mm[ss]]
	regexp.MustCompile(`^(\d{2})(?:(\d{2})(?:(\d{2}))?)?$`),
}

func parseTimeParts(s string) (hour, minute, second int, err error) {
	var parts []string
	for _, re := range timeRegexps {
		parts = re.FindStringSubmatch(s)
		if parts != nil {
			break
		}
	}
	if parts == nil {
		err = fmt.Errorf("soap time: value %q is not in ISO8601 time format", s)
		return
	}

	hour = parseInt(parts[1], &err)
	if len(parts[2]) != 0 {
		minute = parseInt(parts[2], &err)
		if len(parts[3]) != 0 {
			second = parseInt(parts[3], &err)
		}
	}

	if err != nil {
		err = fmt.Errorf("soap time: %q: %v", s, err)
	}

	return
}

// (+|-)hh[[:]mm]
var timezoneRegexp = regexp.MustCompile(`^([+-])(\d{2})(?::?(\d{2}))?$`)

func parseTimezone(s string) (offset int, err error) {
	if s == "Z" {
		return 0, nil
	}
	parts := timezoneRegexp.FindStringSubmatch(s)
	if parts == nil {
		err = fmt.Errorf("soap timezone: value %q is not in ISO8601 timezone format", s)
		return
	}

	offset = parseInt(parts[2], &err) * 3600
	if len(parts[3]) != 0 {
		offset += parseInt(parts[3], &err) * 60
	}
	if parts[1] == "-" {
		offset = -offset
	}

	if err != nil {
		err = fmt.Errorf("soap timezone: %q: %v", s, err)
	}

	return
}

var completeDateTimeZoneRegexp = regexp.MustCompile(`^([^T]+)(?:T([^-+Z]+)(.+)?)?$`)

// splitCompleteDateTimeZone splits date, time and timezone apart from an
// ISO8601 string. It does not ensure that the contents of each part are
// correct, it merely splits on certain delimiters.
// e.g "2010-09-08T12:15:10+0700" => "2010-09-08", "12:15:10", "+0700".
// Timezone can only be present if time is also present.
func splitCompleteDateTimeZone(s string) (dateStr, timeStr, zoneStr string, err error) {
	parts := completeDateTimeZoneRegexp.FindStringSubmatch(s)
	if parts == nil {
		err = fmt.Errorf("soap date/time/zone: value %q is not in ISO8601 datetime format", s)
		return
	}
	dateStr = parts[1]
	timeStr = parts[2]
	zoneStr = parts[3]
	return
}

// marshalUpnpDate marshals time.Time to SOAP "date" type. Note that this converts
// to local time, and discards the time-of-day components.
func marshalUpnpDate(v time.Time) (string, error) {
	return v.In(localLoc).Format("2006-01-02"), nil
}

// unmarshalUpnpDate unmarshals time.Time from SOAP "date" type. This outputs the
// date as midnight in the local time zone.
func unmarshalUpnpDate(s string) (time.Time, error) {
	year, month, day, err := parseDateParts(s)
	if err != nil {
		return time.Time{}, errors.Wrapf(err, "cannot unmarshal Date from '%s'", s)
	}
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, localLoc), nil
}

// marshalUpnpTimeOfDay marshals timeOfDay to the "time" type.
func marshalUpnpTimeOfDay(v timeOfDay) (string, error) {
	d := int64(v.FromMidnight / time.Second)
	hour := d / 3600
	d = d % 3600
	minute := d / 60
	second := d % 60

	return fmt.Sprintf("%02d:%02d:%02d", hour, minute, second), nil
}

// unmarshalUpnpTimeOfDay unmarshals TimeOfDay from the "time" type.
func unmarshalUpnpTimeOfDay(s string) (timeOfDay, error) {
	t, err := unmarshalUpnpTimeOfDayTz(s)
	if err != nil {
		return timeOfDay{}, errors.Wrapf(err, "cannot unmarshal TimeOfDate from '%s'", s)
	} else if t.HasOffset {
		return timeOfDay{}, fmt.Errorf("soap time: value %q contains unexpected timezone", s)
	}
	return t, nil
}

// marshalUpnpTimeOfDayTz marshals timeOfDay to the "time.tz" type.
func marshalUpnpTimeOfDayTz(v timeOfDay) (string, error) {
	d := int64(v.FromMidnight / time.Second)
	hour := d / 3600
	d = d % 3600
	minute := d / 60
	second := d % 60

	tz := ""
	if v.HasOffset {
		if v.Offset == 0 {
			tz = "Z"
		} else {
			offsetMins := v.Offset / 60
			sign := '+'
			if offsetMins < 1 {
				offsetMins = -offsetMins
				sign = '-'
			}
			tz = fmt.Sprintf("%c%02d:%02d", sign, offsetMins/60, offsetMins%60)
		}
	}

	return fmt.Sprintf("%02d:%02d:%02d%s", hour, minute, second, tz), nil
}

// unmarshalUpnpTimeOfDayTz unmarshals TimeOfDay from the "time.tz" type.
func unmarshalUpnpTimeOfDayTz(s string) (tod timeOfDay, err error) {
	zoneIndex := strings.IndexAny(s, "Z+-")
	var timePart string
	var hasOffset bool
	var offset int
	if zoneIndex == -1 {
		hasOffset = false
		timePart = s
	} else {
		hasOffset = true
		timePart = s[:zoneIndex]
		if offset, err = parseTimezone(s[zoneIndex:]); err != nil {
			return
		}
	}

	hour, minute, second, err := parseTimeParts(timePart)
	if err != nil {
		return
	}

	fromMidnight := time.Duration(hour*3600+minute*60+second) * time.Second

	// ISO8601 special case - values up to 24:00:00 are allowed, so using
	// strictly greater-than for the maximum value.
	if fromMidnight > 24*time.Hour || minute >= 60 || second >= 60 {
		return timeOfDay{}, fmt.Errorf("soap time.tz: value %q has value(s) out of range", s)
	}

	return timeOfDay{
		FromMidnight: time.Duration(hour*3600+minute*60+second) * time.Second,
		HasOffset:    hasOffset,
		Offset:       offset,
	}, nil
}

// marshalUpnpDateTime marshals time.Time to SOAP "dateTime" type. Note that this
// converts to local time.
func marshalUpnpDateTime(v time.Time) (string, error) {
	return v.In(localLoc).Format("2006-01-02T15:04:05"), nil
}

// unmarshalUpnpDateTime unmarshals time.Time from the SOAP "dateTime" type. This
// returns a value in the local timezone.
func unmarshalUpnpDateTime(s string) (result time.Time, err error) {
	dateStr, timeStr, zoneStr, err := splitCompleteDateTimeZone(s)
	if err != nil {
		return time.Time{}, errors.Wrapf(err, "cannot unmarshal DateTime from '%s'", s)
	}

	if len(zoneStr) != 0 {
		err = fmt.Errorf("soap datetime: unexpected timezone in %q", s)
		return
	}

	year, month, day, err := parseDateParts(dateStr)
	if err != nil {
		return
	}

	var hour, minute, second int
	if len(timeStr) != 0 {
		hour, minute, second, err = parseTimeParts(timeStr)
		if err != nil {
			return
		}
	}

	result = time.Date(year, time.Month(month), day, hour, minute, second, 0, localLoc)
	return
}

// marshalUpnpDateTimeTz marshals time.Time to SOAP "dateTime.tz" type.
func marshalUpnpDateTimeTz(v time.Time) (string, error) {
	return v.Format("2006-01-02T15:04:05-07:00"), nil
}

// unmarshalUpnpDateTimeTz unmarshals time.Time from the SOAP "dateTime.tz" type.
// This returns a value in the local timezone when the timezone is unspecified.
func unmarshalUpnpDateTimeTz(s string) (result time.Time, err error) {
	dateStr, timeStr, zoneStr, err := splitCompleteDateTimeZone(s)
	if err != nil {
		return time.Time{}, errors.Wrapf(err, "cannot unmarshal DateTimeTz from '%s'", s)
	}

	year, month, day, err := parseDateParts(dateStr)
	if err != nil {
		return
	}

	var hour, minute, second int
	var location *time.Location = localLoc
	if len(timeStr) != 0 {
		hour, minute, second, err = parseTimeParts(timeStr)
		if err != nil {
			return
		}
		if len(zoneStr) != 0 {
			var offset int
			offset, err = parseTimezone(zoneStr)
			if offset == 0 {
				location = time.UTC
			} else {
				location = time.FixedZone("", offset)
			}
		}
	}

	result = time.Date(year, time.Month(month), day, hour, minute, second, 0, location)
	return
}

// marshalUpnpBoolean marshals bool to SOAP "boolean" type.
func marshalUpnpBoolean(v bool) (string, error) {
	if v {
		return "1", nil
	}
	return "0", nil
}

// unmarshalUpnpBoolean unmarshals bool from the SOAP "boolean" type.
func unmarshalUpnpBoolean(s string) (bool, error) {
	switch s {
	case "0", "false", "no":
		return false, nil
	case "1", "true", "yes":
		return true, nil
	}
	return false, fmt.Errorf("cannot unmarshal Boolean from '%s'", s)
}

// marshalUpnpBinBase64 marshals []byte to SOAP "bin.base64" type.
func marshalUpnpBinBase64(v []byte) (string, error) {
	return base64.StdEncoding.EncodeToString(v), nil
}

// unmarshalUpnpBinBase64 unmarshals []byte from the SOAP "bin.base64" type.
func unmarshalUpnpBinBase64(s string) ([]byte, error) {
	v, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return []byte{}, errors.Wrapf(err, "cannot unmarshal BinBase64 from '%s'", s)
	}
	return v, nil
}

// marshalUpnpBinHex marshals []byte to SOAP "bin.hex" type.
func marshalUpnpBinHex(v []byte) (string, error) {
	return hex.EncodeToString(v), nil
}

// unmarshalUpnpBinHex unmarshals []byte from the SOAP "bin.hex" type.
func unmarshalUpnpBinHex(s string) ([]byte, error) {
	v, err := hex.DecodeString(s)
	if err != nil {
		return []byte{}, errors.Wrapf(err, "cannot unmarshal BinHex from '%s'", s)
	}
	return v, nil
}

// marshalUpnpURI marshals *url.URL to SOAP "uri" type.
func marshalUpnpURI(v *url.URL) (string, error) {
	return v.String(), nil
}

// unmarshalUpnpURI unmarshals *url.URL from the SOAP "uri" type.
func unmarshalUpnpURI(s string) (*url.URL, error) {
	v, err := url.Parse(s)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot unmarshal URI from '%s'", s)
	}
	return v, nil
}
