package apache_log_counter

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"
	"time"
)

const (
	TimeFormat = "02/Jan/2006:15:04:05"
	GeoIpUrl   = "https://geoipapi-216405.appspot.com/"
)

type (
	ApacheLogCounter struct {
		Filter Filter
		GeoIp  GeoIp
	}

	Filter struct {
		Host       string `json:"host"`
		RemoteHost string `json:"remote_host"`
		StartTime  string `json:"start_time"`
		EndTime    string `json:"end_time"`
		Status     int    `json:"status"`
		Method     string `json:"method"`
		RequestURI string `json:"request_uri"`
		Request    string `json:"request"`
	}

	GeoIp struct {
		IsoCode string `json:"iso_code"`
		Country string `json:"country"`
		Error   error  `json:"error"`
	}

	Counters []Counter
	Counter  struct {
		Key   string
		Value int
	}
)

func NewApacheLogCounter() *ApacheLogCounter {
	return new(ApacheLogCounter)
}

func (alc *ApacheLogCounter) ParseFilter(s string) error {
	if s == "" {
		return nil
	}
	err := json.Unmarshal([]byte(s), &alc.Filter)
	return err
}

func (alc *ApacheLogCounter) GetStartTime() (time.Time, bool, error) {
	if alc.Filter.StartTime == "" {
		return time.Time{}, false, nil
	}
	t, err := time.ParseInLocation(TimeFormat, alc.Filter.StartTime, time.Local)
	return t, true, err
}

func (alc *ApacheLogCounter) GetEndTime() (time.Time, bool, error) {
	if alc.Filter.EndTime == "" {
		return time.Time{}, false, nil
	}
	t, err := time.ParseInLocation(TimeFormat, alc.Filter.EndTime, time.Local)
	return t, true, err
}

func SortDesc(counts map[string]int) Counters {
	c := make(Counters, len(counts))
	i := 0
	for k, v := range counts {
		c[i] = Counter{k, v}
		i++
	}
	sort.SliceStable(c, func(i, j int) bool {
		return c[i].Value > c[j].Value
	})
	return c
}

func (alc *ApacheLogCounter) FetchRemoteHostCountry(ip string) (string, string, error) {
	resp, err := http.Get(GeoIpUrl + ip)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(alc.GeoIp); err != nil {
		return "", "", err
	}

	return alc.GeoIp.IsoCode, alc.GeoIp.Country, alc.GeoIp.Error
}

func (alc *ApacheLogCounter) FilteringHost(host string) bool {
	if alc.Filter.Host != "" && (alc.Filter.Host != host) {
		return true
	}
	return false
}

func (alc *ApacheLogCounter) FilteringRemoteHost(remoteHost string) bool {
	if alc.Filter.RemoteHost != "" && (alc.Filter.RemoteHost != remoteHost) {
		return true
	}
	return false
}

func (alc *ApacheLogCounter) FilteringStatus(status int) bool {
	if alc.Filter.Status != 0 && (alc.Filter.Status != status) {
		return true
	}
	return false
}

func (alc *ApacheLogCounter) FilteringMethod(method string) bool {
	if alc.Filter.Method != "" && (strings.ToUpper(alc.Filter.Method) != method) {
		return true
	}
	return false
}

func (alc *ApacheLogCounter) FilteringRequestURI(requestURI string) bool {
	if alc.Filter.RequestURI != "" && (alc.Filter.RequestURI != requestURI) {
		return true
	}
	return false
}

func (alc *ApacheLogCounter) FilteringRequest(request string) bool {
	if alc.Filter.Request != "" && (alc.Filter.Request != request) {
		return true
	}
	return false
}

func (alc *ApacheLogCounter) FilteringStartTime(startTime time.Time) (bool, error) {
	t, ok, err := alc.GetStartTime()
	if !ok {
		return false, nil
	} else if err != nil {
		return true, err
	}

	b := startTime.Before(t)
	return b, nil
}

func (alc *ApacheLogCounter) FilteringEndTime(endTime time.Time) (bool, error) {
	t, ok, err := alc.GetEndTime()
	if !ok {
		return false, nil
	} else if err != nil {
		return true, err
	}

	b := endTime.After(t)
	return b, nil
}
