package schemer

import (
	"math/rand"
	"net/url"

	"github.com/gookit/config/v2"
	"github.com/gookit/config/v2/yamlv3"
	"github.com/rs/zerolog/log"
	"github.com/ungerik/go-dry"

	"github.com/goccy/go-yaml"
)

type Scheme struct {
	// Mighty used when we want for non-pre-defined fields
	Mighty *config.Config

	Id  int    `yaml:"id,omitempty"`
	Uid string `yaml:"uid,omitempty"`

	Ctrl *Ctrl `yaml:"ctrl,omitempty"`

	Mapper *Mapper `yaml:"mapper,omitempty"`

	// Kwargs values of Selectors
	Kwargs *Kwargs `yaml:"kwargs,omitempty"`
}

type Ctrl struct {
	Humanized bool `yaml:"humanized,omitempty"`
	Hijack    bool `yaml:"hijack,omitempty"`
}

type Mapper struct {
	Home      string   `yaml:"home,omitempty"`
	NavChains []string `yaml:"nav_chains,omitempty"`
	Popovers  []string `yaml:"popovers,omitempty"`
	Cookies   []string `yaml:"cookies,omitempty"`

	SortBy *Dropdown `yaml:"sort_by,omitempty"`

	Inputs []string `yaml:"inputs,omitempty"`
	Submit string   `yaml:"submit,omitempty"`

	Filters []string `yaml:"filters,omitempty"`

	HasResults   []string `yaml:"has_results,omitempty"`
	NoResults    []string `yaml:"no_results,omitempty"`
	ResultsCount string   `yaml:"results_count,omitempty"`

	NextPage string `yaml:"next_page,omitempty"`
	PageNum  string `yaml:"page_num,omitempty"`
}

type Kwargs struct {
	MaxScrollTime int `yaml:"max_scroll_time,omitempty"`

	MaxPages      int     `yaml:"max_pages,omitempty"`
	NextPageIndex int     `yaml:"next_page_index,omitempty"`
	PageInterval  float64 `yaml:"page_interval,omitempty"`

	ResultsCount ResultsCount `yaml:"results_count,omitempty"`

	ResponseParseScript string `yaml:"response_parse_script,omitempty"`
}

type ResultsCount struct {
	// Index if the selector got more than 1 elems
	Index int
	// attr
	Attr      string
	AttrSep   string
	AttrIndex int

	// CharsAllowed is used when there exists thousand separator i.e. `"." or ","`
	CharsAllowed string
}

type Dropdown struct {
	Trigger string
	Item    string
}

func NewDropdown(t, s string) *Dropdown {
	return &Dropdown{t, s}
}

func NewScheme(raw []byte) (*Scheme, error) {
	scheme := Scheme{}

	err := yaml.Unmarshal(raw, &scheme)
	if err != nil {
		return nil, err
	}

	cfg, err := yaml2Config(raw)
	if err != nil {
		return nil, err
	}

	scheme.Mighty = cfg

	if scheme.Uid == "" {
		v, _ := url.Parse(scheme.Mapper.Home)
		scheme.Uid = v.Host
	}

	if scheme.Kwargs.MaxPages == 0 {
		scheme.Kwargs.MaxPages = rand.Intn(6) + 20 // nolint: gomnd
	}

	return &scheme, err
}

func NewSchemeFromFile(filename string) (*Scheme, error) {
	raw, err := dry.FileGetBytes(filename)
	if err != nil {
		log.Error().Str("filename", filename).Msg("cannot load startup")
	}

	return NewSchemeFromByte(raw)
}

func NewSchemeFromByte(raw []byte) (*Scheme, error) {
	scheme, err := NewScheme(raw)
	if err != nil {
		return nil, err
	}

	return scheme, nil
}

func yaml2Config(raw []byte, srcConfigArr ...*config.Config) (*config.Config, error) {
	var cfg *config.Config

	if len(srcConfigArr) != 0 {
		cfg = srcConfigArr[0]
	} else {
		cfg = config.New("")
	}

	cfg.AddDriver(yamlv3.Driver)

	err := cfg.LoadSources(config.Yaml, raw)

	return cfg, err
}
