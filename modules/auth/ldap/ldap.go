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

package ldap

import (
	"crypto/tls"
	"fmt"
	"strings"

	log "gopkg.in/clog.v1"
	"gopkg.in/ldap.v2"
)

type SecurityProtocol int

// Note: new type must be added at the end of list to maintain compatibility.
const (
	SECURITY_PROTOCOL_UNENCRYPTED SecurityProtocol = iota
	SECURITY_PROTOCOL_LDAPS
	SECURITY_PROTOCOL_START_TLS
)

// Basic LDAP authentication service
type Source struct {
	Name              string // canonical name (ie. corporate.ad)
	Host              string // LDAP host
	Port              int    // port number
	SecurityProtocol  SecurityProtocol
	SkipVerify        bool
	BindDN            string // DN to bind with
	BindPassword      string // Bind DN password
	UserBase          string // Base search path for users
	UserDN            string // Template for the DN of the user for simple auth
	AttributeUsername string // Username attribute
	AttributeName     string // First name attribute
	AttributeSurname  string // Surname attribute
	AttributeMail     string // E-mail attribute
	AttributesInBind  bool   // fetch attributes in bind context (not user)
	Filter            string // Query filter to validate entry
	AdminFilter       string // Query filter to check if user is admin
	Enabled           bool   // if this source is disabled
}

func (ls *Source) sanitizedUserQuery(username string) (string, bool) {
	// See http://tools.ietf.org/search/rfc4515
	badCharacters := "\x00()*\\"
	if strings.ContainsAny(username, badCharacters) {
		log.Trace("Username contains invalid query characters: %s", username)
		return "", false
	}

	return fmt.Sprintf(ls.Filter, username), true
}

func (ls *Source) sanitizedUserDN(username string) (string, bool) {
	// See http://tools.ietf.org/search/rfc4514: "special characters"
	badCharacters := "\x00()*\\,='\"#+;<>"
	if strings.ContainsAny(username, badCharacters) || strings.HasPrefix(username, " ") || strings.HasSuffix(username, " ") {
		log.Trace("Username contains invalid query characters: %s", username)
		return "", false
	}

	return fmt.Sprintf(ls.UserDN, username), true
}

func (ls *Source) findUserDN(l *ldap.Conn, name string) (string, bool) {
	log.Trace("Search for LDAP user: %s", name)
	if ls.BindDN != "" && ls.BindPassword != "" {
		err := l.Bind(ls.BindDN, ls.BindPassword)
		if err != nil {
			log.Trace("Failed to bind as BindDN '%s': %v", ls.BindDN, err)
			return "", false
		}
		log.Trace("Bound as BindDN: %s", ls.BindDN)
	} else {
		log.Trace("Proceeding with anonymous LDAP search")
	}

	// A search for the user.
	userFilter, ok := ls.sanitizedUserQuery(name)
	if !ok {
		return "", false
	}

	log.Trace("Searching for DN using filter '%s' and base '%s'", userFilter, ls.UserBase)
	search := ldap.NewSearchRequest(
		ls.UserBase, ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0,
		false, userFilter, []string{}, nil)

	// Ensure we found a user
	sr, err := l.Search(search)
	if err != nil || len(sr.Entries) < 1 {
		log.Trace("Failed search using filter '%s': %v", userFilter, err)
		return "", false
	} else if len(sr.Entries) > 1 {
		log.Trace("Filter '%s' returned more than one user", userFilter)
		return "", false
	}

	userDN := sr.Entries[0].DN
	if userDN == "" {
		log.Error(4, "LDAP search was successful, but found no DN!")
		return "", false
	}

	return userDN, true
}

func dial(ls *Source) (*ldap.Conn, error) {
	log.Trace("Dialing LDAP with security protocol '%v' without verifying: %v", ls.SecurityProtocol, ls.SkipVerify)

	tlsCfg := &tls.Config{
		ServerName:         ls.Host,
		InsecureSkipVerify: ls.SkipVerify,
	}
	if ls.SecurityProtocol == SECURITY_PROTOCOL_LDAPS {
		return ldap.DialTLS("tcp", fmt.Sprintf("%s:%d", ls.Host, ls.Port), tlsCfg)
	}

	conn, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", ls.Host, ls.Port))
	if err != nil {
		return nil, fmt.Errorf("Dial: %v", err)
	}

	if ls.SecurityProtocol == SECURITY_PROTOCOL_START_TLS {
		if err = conn.StartTLS(tlsCfg); err != nil {
			conn.Close()
			return nil, fmt.Errorf("StartTLS: %v", err)
		}
	}

	return conn, nil
}

func bindUser(l *ldap.Conn, userDN, passwd string) error {
	log.Trace("Binding with userDN: %s", userDN)
	err := l.Bind(userDN, passwd)
	if err != nil {
		log.Trace("LDAP authentication failed for '%s': %v", userDN, err)
		return err
	}
	log.Trace("Bound successfully with userDN: %s", userDN)
	return err
}

// searchEntry : search an LDAP source if an entry (name, passwd) is valid and in the specific filter
func (ls *Source) SearchEntry(name, passwd string, directBind bool) (string, string, string, string, bool, bool) {
	// See https://tools.ietf.org/search/rfc4513#section-5.1.2
	if len(passwd) == 0 {
		log.Trace("authentication failed for '%s' with empty password")
		return "", "", "", "", false, false
	}
	l, err := dial(ls)
	if err != nil {
		log.Error(4, "LDAP connect failed for '%s': %v", ls.Host, err)
		ls.Enabled = false
		return "", "", "", "", false, false
	}
	defer l.Close()

	var userDN string
	if directBind {
		log.Trace("LDAP will bind directly via UserDN template: %s", ls.UserDN)

		var ok bool
		userDN, ok = ls.sanitizedUserDN(name)
		if !ok {
			return "", "", "", "", false, false
		}
	} else {
		log.Trace("LDAP will use BindDN")

		var found bool
		userDN, found = ls.findUserDN(l, name)
		if !found {
			return "", "", "", "", false, false
		}
	}

	if directBind || !ls.AttributesInBind {
		// binds user (checking password) before looking-up attributes in user context
		err = bindUser(l, userDN, passwd)
		if err != nil {
			return "", "", "", "", false, false
		}
	}

	userFilter, ok := ls.sanitizedUserQuery(name)
	if !ok {
		return "", "", "", "", false, false
	}

	log.Trace("Fetching attributes '%v', '%v', '%v', '%v' with filter '%s' and base '%s'", ls.AttributeUsername, ls.AttributeName, ls.AttributeSurname, ls.AttributeMail, userFilter, userDN)
	search := ldap.NewSearchRequest(
		userDN, ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false, userFilter,
		[]string{ls.AttributeUsername, ls.AttributeName, ls.AttributeSurname, ls.AttributeMail},
		nil)

	sr, err := l.Search(search)
	if err != nil {
		log.Error(4, "LDAP search failed: %v", err)
		return "", "", "", "", false, false
	} else if len(sr.Entries) < 1 {
		if directBind {
			log.Error(4, "User filter inhibited user login")
		} else {
			log.Error(4, "LDAP search failed: 0 entries")
		}

		return "", "", "", "", false, false
	}

	username := sr.Entries[0].GetAttributeValue(ls.AttributeUsername)
	firstname := sr.Entries[0].GetAttributeValue(ls.AttributeName)
	surname := sr.Entries[0].GetAttributeValue(ls.AttributeSurname)
	mail := sr.Entries[0].GetAttributeValue(ls.AttributeMail)

	isAdmin := false
	if len(ls.AdminFilter) > 0 {
		log.Trace("Checking admin with filter '%s' and base '%s'", ls.AdminFilter, userDN)
		search = ldap.NewSearchRequest(
			userDN, ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false, ls.AdminFilter,
			[]string{ls.AttributeName},
			nil)

		sr, err = l.Search(search)
		if err != nil {
			log.Error(4, "LDAP admin search failed: %v", err)
		} else if len(sr.Entries) < 1 {
			log.Error(4, "LDAP admin search failed: 0 entries")
		} else {
			isAdmin = true
		}
	}

	if !directBind && ls.AttributesInBind {
		// binds user (checking password) after looking-up attributes in BindDN context
		err = bindUser(l, userDN, passwd)
		if err != nil {
			return "", "", "", "", false, false
		}
	}

	return username, firstname, surname, mail, isAdmin, true
}
