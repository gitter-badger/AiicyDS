// Copyright 2017 The Aiicy Team.
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package auth

import (
	"reflect"
	"strings"

	macaron "gopkg.in/macaron.v1"

	"github.com/Unknwon/com"
	"github.com/go-macaron/binding"
)

func IsAPIPath(url string) bool {
	return strings.HasPrefix(url, "/api/")
}

type Form interface {
	binding.Validator
}

// AssignForm assign form values back to the template data.
func AssignForm(form interface{}, data map[string]interface{}) {
	typ := reflect.TypeOf(form)
	val := reflect.ValueOf(form)

	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
		val = val.Elem()
	}

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		fieldName := field.Tag.Get("form")
		// Allow ignored fields in the struct
		if fieldName == "-" {
			continue
		} else if len(fieldName) == 0 {
			fieldName = com.ToSnakeCase(field.Name)
		}

		data[fieldName] = val.Field(i).Interface()
	}
}

func getRuleBody(field reflect.StructField, prefix string) string {
	for _, rule := range strings.Split(field.Tag.Get("binding"), ";") {
		if strings.HasPrefix(rule, prefix) {
			return rule[len(prefix) : len(rule)-1]
		}
	}
	return ""
}

func GetSize(field reflect.StructField) string {
	return getRuleBody(field, "Size(")
}

func GetMinSize(field reflect.StructField) string {
	return getRuleBody(field, "MinSize(")
}

func GetMaxSize(field reflect.StructField) string {
	return getRuleBody(field, "MaxSize(")
}

func GetInclude(field reflect.StructField) string {
	return getRuleBody(field, "Include(")
}

func validate(errs binding.Errors, data map[string]interface{}, f Form, l macaron.Locale) binding.Errors {
	if errs.Len() == 0 {
		return errs
	}

	data["HasError"] = true
	AssignForm(f, data)

	typ := reflect.TypeOf(f)
	val := reflect.ValueOf(f)

	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
		val = val.Elem()
	}

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		fieldName := field.Tag.Get("form")
		// Allow ignored fields in the struct
		if fieldName == "-" {
			continue
		}

		if errs[0].FieldNames[0] == field.Name {
			data["Err_"+field.Name] = true

			trName := field.Tag.Get("locale")
			if len(trName) == 0 {
				trName = l.Tr("form." + field.Name)
			} else {
				trName = l.Tr(trName)
			}

			switch errs[0].Classification {
			case binding.ERR_REQUIRED:
				data["ErrorMsg"] = trName + l.Tr("form.require_error")
			case binding.ERR_ALPHA_DASH:
				data["ErrorMsg"] = trName + l.Tr("form.alpha_dash_error")
			case binding.ERR_ALPHA_DASH_DOT:
				data["ErrorMsg"] = trName + l.Tr("form.alpha_dash_dot_error")
			case binding.ERR_SIZE:
				data["ErrorMsg"] = trName + l.Tr("form.size_error", GetSize(field))
			case binding.ERR_MIN_SIZE:
				data["ErrorMsg"] = trName + l.Tr("form.min_size_error", GetMinSize(field))
			case binding.ERR_MAX_SIZE:
				data["ErrorMsg"] = trName + l.Tr("form.max_size_error", GetMaxSize(field))
			case binding.ERR_EMAIL:
				data["ErrorMsg"] = trName + l.Tr("form.email_error")
			case binding.ERR_URL:
				data["ErrorMsg"] = trName + l.Tr("form.url_error")
			case binding.ERR_INCLUDE:
				data["ErrorMsg"] = trName + l.Tr("form.include_error", GetInclude(field))
			default:
				data["ErrorMsg"] = l.Tr("form.unknown_error") + " " + errs[0].Classification
			}
			return errs
		}
	}
	return errs
}
