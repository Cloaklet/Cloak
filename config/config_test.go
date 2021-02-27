package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
)

type configTestSuite struct {
	suite.Suite
	cfg *Configurator
}

func (s *configTestSuite) SetupSuite() {
	var err error
	s.cfg, err = NewConfigurator("./config_test_sample.ini")
	s.Require().NoError(err)
	s.Require().NotNil(s.cfg, "NewConfigurator() should return a new instance of Configurator")
	s.Require().NotEmpty(s.cfg.filePath)
}

func (s *configTestSuite) Test_01_Load() {
	var callbackFiredCount int
	var expectedCallbackFireCount int
	s.cfg.SetCallback("root_key", func(v string) error {
		s.Require().EqualValues("root_value", v)
		callbackFiredCount += 1
		return nil
	})
	expectedCallbackFireCount += 1
	s.cfg.SetCallback("root_key2", func(v string) error {
		s.Require().EqualValues("root_value 2", v)
		callbackFiredCount += 1
		return nil
	})
	expectedCallbackFireCount += 1
	s.cfg.SetCallback("section1.subsection1.subsubsec1.my_key", func(v string) error {
		s.Require().EqualValues("value", v)
		callbackFiredCount += 1
		return nil
	})
	expectedCallbackFireCount += 1
	s.cfg.SetCallback("section1.subsection1.subsec1_key", func(v string) error {
		s.Require().EqualValues("some value", v)
		callbackFiredCount += 1
		return nil
	})
	expectedCallbackFireCount += 1
	s.Require().NoError(s.cfg.Load())

	s.Require().EqualValues(expectedCallbackFireCount, callbackFiredCount)
	s.Require().NotNil(s.cfg.ini)
	s.Require().NotEmpty(s.cfg.data)
}

func (s *configTestSuite) Test_02_Load_InvalidPath() {
	cfg, err := NewConfigurator("/some/non/existing/path/invalid.ini")
	s.Require().Error(err)
	s.Require().Nil(cfg)
}
func (s *configTestSuite) Test_03_LoadEmpty_Set() {

	cfg, err := NewConfigurator("./non-existing.ini")
	s.Require().NoError(err)
	s.Require().NotNil(cfg)
	_, err = os.Stat("./non-existing.ini")
	s.Require().Error(err)
	s.Require().True(os.IsNotExist(err))

	s.Require().NoError(cfg.Load())
	defer func() {
		s.NoError(os.Remove("./non-existing.ini"))
	}()

	s.Require().EqualValues(0, len(cfg.data))
	info, err := os.Stat("./non-existing.ini")
	s.Require().NoError(err)
	s.Require().False(info.IsDir())

	// Set values
	cfg.SetCallback("root_key", func(v string) error {
		s.Require().EqualValues("root_value", v)
		return nil
	})
	s.Require().NoError(cfg.Set("root_key", "root_value"))

	// Inspect INI file content
	iniContent, err := ioutil.ReadFile("./non-existing.ini")
	s.Require().NoError(err)
	s.Require().Contains(string(iniContent), "=")
	s.Require().Contains(string(iniContent), "root_value")
	fmt.Print(string(iniContent))
}

func TestConfigurator(t *testing.T) {
	suite.Run(t, new(configTestSuite))
}
