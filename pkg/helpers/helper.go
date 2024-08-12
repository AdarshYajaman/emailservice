package helpers

import (
	"fmt"
	"log"
	"path/filepath"
	"text/template"

	"citi.com/179563_genesis_mail/pkg/config"
	"github.com/spf13/viper"
)

func ReadConfigFile() *config.ApplicationProperties {
	appProps := config.ApplicationProperties{}

	viper.SetConfigName("application")
	viper.AddConfigPath("./pkg/config/")
	viper.SetConfigType("properties")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
	err = viper.Unmarshal(&appProps)
	if err != nil {
		log.Fatalf("unable to decode into struct, %v", err)
	}
	return &appProps

}

func CreateTemplateCache() (map[string]*template.Template, error) {
	myCache := map[string]*template.Template{}

	matches, err := filepath.Glob("./templates/*.page.tmpl")
	if err != nil {
		return myCache, err
	}

	for _, value := range matches {
		name := filepath.Base(value)

		ts, err := template.New(name).ParseFiles(value)
		if err != nil {
			return myCache, err
		}

		layouts, err := filepath.Glob("./templates/*.layout.tmpl")

		if err != nil {
			return myCache, err
		}

		if len(layouts) > 0 {
			ts, err = ts.ParseGlob("./templates/*.layout.tmpl")
			if err != nil {
				return myCache, err
			}
		}

		myCache[name] = ts

	}

	return myCache, nil

}
