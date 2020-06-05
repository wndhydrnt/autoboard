package v1

import "strings"

type FormatMapper struct {
}

func (fm *FormatMapper) Find(metricName string) string {
	if strings.HasSuffix(metricName, "_bytes") {
		return "decbytes"
	}

	if strings.HasSuffix(metricName, "_duration_seconds") {
		return "s"
	}

	if strings.HasSuffix(metricName, "_interval_seconds") {
		return "s"
	}

	if strings.HasSuffix(metricName, "_range_seconds") {
		return "s"
	}

	if strings.HasSuffix(metricName, "_cleanup_seconds") {
		return "s"
	}

	if strings.HasSuffix(metricName, "_seconds") {
		return "dateTimeAsIso"
	}

	if strings.HasSuffix(metricName, "_timestamp") {
		return "dateTimeAsIso"
	}

	if strings.HasSuffix(metricName, "_time") {
		return "dateTimeAsIso"
	}

	return "short"
}

func (fm *FormatMapper) FindRange(metricName string) string {
	if strings.HasSuffix(metricName, "_bytes_total") || strings.HasSuffix(metricName, "_bytes") {
		return "Bps"
	}

	if strings.HasSuffix(metricName, "_requests_total") {
		return "reqps"
	}

	return "short"
}

var DefaultFormatMapper = FormatMapper{}

func FindFormat(metricName string) string {
	return DefaultFormatMapper.Find(metricName)
}

func FindRangeFormat(metricName string) string {
	return DefaultFormatMapper.FindRange(metricName)
}
