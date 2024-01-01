package schemer

import (
	"testing"

	"github.com/k0kubun/pp/v3"
	"github.com/stretchr/testify/suite"
	"github.com/ungerik/go-dry"
)

type StartupSuite struct {
	suite.Suite
	rawBytes []byte
}

func TestStartup(t *testing.T) {
	suite.Run(t, new(StartupSuite))
}

func (s *StartupSuite) SetupSuite() {
	raw, err := dry.FileGetBytes("scheme.yaml")
	s.Nil(err)
	s.rawBytes = raw
}

func (s *StartupSuite) TearDownSuite() {
}

func (s *StartupSuite) Test_01_startupAsStruct() {
	st, err := NewScheme(s.rawBytes)
	s.Nil(err)

	pp.Println(st.Mapper.Home)
}

func (s *StartupSuite) Test_02_startupAsConfig() {
	st, err := NewScheme(s.rawBytes)
	s.Nil(err)

	pp.Println(st.Kwargs)

	cfg := st.Mighty

	s.Equal(st.Mapper.Home, cfg.String("home"), "home")
	s.Equal(st.Id, cfg.Int("id"), "id")
	s.Equal(st.Uid, "www.example.com")
	s.Equal(`ul+div[class^=loading]@@@--@@@Loading...`, cfg.String("selectors.loading"), "config-a.b")
	s.Equal(st.Mapper.Inputs, []string{"input.name", "input.location"})
}
