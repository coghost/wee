package wee

import (
	"encoding/json"
	"fmt"
)

const (
	defaultURL = "https://ifconfig.co/json"
)

type IPInfo struct {
	IP         string  `json:"ip"`
	IPDecimal  int64   `json:"ip_decimal"`
	Country    string  `json:"country"`
	CountryIso string  `json:"country_iso"`
	CountryEu  bool    `json:"country_eu"`
	RegionName string  `json:"region_name"`
	RegionCode string  `json:"region_code"`
	ZipCode    string  `json:"zip_code"`
	City       string  `json:"city"`
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
	TimeZone   string  `json:"time_zone"`
	Asn        string  `json:"asn"`
	AsnOrg     string  `json:"asn_org"`
	UserAgent  struct {
		Product  string `json:"product"`
		Version  string `json:"version"`
		Comment  string `json:"comment"`
		RawValue string `json:"raw_value"`
	} `json:"user_agent"`
}

// CurlIPFromIfconfigCO loads `https://ifconfig.co/json`
// return the Unmarshal data, error
func CurlIPFromIfconfigCO(bot *Bot) (*IPInfo, error) {
	cleanUprequired := false

	if bot == nil {
		bot = NewBotHeadless()
		cleanUprequired = true
	}

	if cleanUprequired {
		defer bot.Cleanup()
	}

	if err := bot.Open(defaultURL); err != nil {
		return nil, fmt.Errorf("cannot open url: %s, %w", defaultURL, err)
	}

	raw, err := bot.ElemAttr(`body>pre`, WithTimeout(MediumToSec))
	if err != nil {
		return nil, fmt.Errorf("cannot get ip info: %w", err)
	}

	var co *IPInfo
	err = json.Unmarshal([]byte(raw), &co)

	return co, err
}
