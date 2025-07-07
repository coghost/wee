package wee

import (
	"testing"

	"github.com/k0kubun/pp/v3"
	"github.com/stretchr/testify/assert"
)

func TestCurl(t *testing.T) {
	assert := assert.New(t)

	got, err := CurlIPFromIfconfigCO(nil)
	assert.Nil(err)

	pp.Println(got)
}
