package v1

import "bytes"

type FormatMapper struct {
}

func (fm *FormatMapper) Find(metricName []byte) string {
	if bytes.HasSuffix(metricName, []byte("_bytes")) {
		return "decbytes"
	}

	if bytes.HasSuffix(metricName, []byte("_duration_seconds")) {
		return "s"
	}

	if bytes.HasSuffix(metricName, []byte("_interval_seconds")) {
		return "s"
	}

	if bytes.HasSuffix(metricName, []byte("_range_seconds")) {
		return "s"
	}

	if bytes.HasSuffix(metricName, []byte("_cleanup_seconds")) {
		return "s"
	}

	if bytes.HasSuffix(metricName, []byte("_seconds")) {
		return "dateTimeAsIso"
	}

	if bytes.HasSuffix(metricName, []byte("_timestamp")) {
		return "dateTimeAsIso"
	}

	if bytes.HasSuffix(metricName, []byte("_time")) {
		return "dateTimeAsIso"
	}

	return "short"
}

func (fm *FormatMapper) FindRange(metricName []byte) string {
	if bytes.HasSuffix(metricName, []byte("_bytes_total")) || bytes.HasSuffix(metricName, []byte("_bytes")) {
		return "Bps"
	}

	if bytes.HasSuffix(metricName, []byte("_requests_total")) {
		return "reqps"
	}

	return "short"
}

var DefaultFormatMapper = FormatMapper{}

func FindFormat(metricName []byte) string {
	return DefaultFormatMapper.Find(metricName)
}

func FindRangeFormat(metricName []byte) string {
	return DefaultFormatMapper.FindRange(metricName)
}
