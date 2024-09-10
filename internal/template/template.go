// Copyright 2024 Aleksei Grigorev
// https://aleksvgrig.com, https://github.com/AlekseiGrigorev, aleksvgrig@gmail.com.
// Package define interfaces, structures and functions for working with templates.
// Template manager replaces placeholders in the template and return result text.
package template

import (
	"fmt"
	"os"
	"strings"

	"github.com/AlekseiGrigorev/ydloader/internal/trace"
)

// TemplateManager define base struct for template manager
// Template manager replaces placeholders in the template and return result text.
type TemplateManager struct {
	template string
}

// SetTemplate read template from file
func (tm *TemplateManager) SetTemplate(templatePath string) error {
	if _, err := os.Stat(templatePath); err != nil {
		fmt.Println(err, trace.GetTrace())
		return err
	}
	b, err := os.ReadFile(templatePath)
	if err != nil {
		fmt.Println(err, trace.GetTrace())
		return err
	}
	tm.template = string(b)
	return nil
}

// Process template replaces placeholders and return result text
func (tm *TemplateManager) Process(params map[string]string) string {
	if tm.template == "" {
		return ""
	}
	s := tm.template
	for paramsKey, paramsValue := range params {
		s = strings.Replace(s, paramsKey, paramsValue, -1)
	}
	return s
}
