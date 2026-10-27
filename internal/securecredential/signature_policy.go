package securecredential

import (
	"bytes"
	"encoding/xml"
	"regexp"
	"strconv"
	"strings"
)

const (
	trustedTeamID      = "3PVV8DQPL6"
	trustedIdentifier  = "io.riido.cli-auth-helper"
	trustedAccessGroup = trustedTeamID + "." + trustedIdentifier
)

var (
	codeFlags          = regexp.MustCompile(`(?:^| )flags=0x([0-9a-fA-F]+)(?:\(| |$)`)
	unsafeEntitlements = []string{"get-task-allow", "com.apple.security.get-task-allow", "com.apple.security.cs.disable-library-validation", "com.apple.security.cs.allow-unsigned-executable-memory", "com.apple.security.cs.allow-dyld-environment-variables", "com.apple.security.cs.allow-jit", "com.apple.security.cs.disable-executable-page-protection"}
)

type plistNode struct {
	XMLName  xml.Name
	Text     string      `xml:",chardata"`
	Children []plistNode `xml:",any"`
}

func checkSignature(metadata string, entitlements []byte) error {
	team, id := "", ""
	runtime := false
	for line := range strings.SplitSeq(metadata, "\n") {
		if after, ok := strings.CutPrefix(line, "TeamIdentifier="); ok {
			team = after
		}
		if after, ok := strings.CutPrefix(line, "Identifier="); ok {
			id = after
		}
		if strings.HasPrefix(line, "CodeDirectory ") {
			if match := codeFlags.FindStringSubmatch(line); len(match) == 2 {
				flags, err := strconv.ParseUint(match[1], 16, 32)
				runtime = err == nil && flags&0x10000 != 0 && flags&0x2 == 0
			}
		}
	}
	if team != trustedTeamID || id != trustedIdentifier || !runtime {
		return ErrUnavailable
	}
	values, err := entitlementValues(entitlements)
	if err != nil {
		return err
	}
	for _, key := range unsafeEntitlements {
		if v, ok := values[key]; ok && v.XMLName.Local != "false" {
			return ErrUnavailable
		}
	}
	app, ok := values["application-identifier"]
	if !ok {
		app = values["com.apple.application-identifier"]
	}
	if app.XMLName.Local != "string" || strings.TrimSpace(app.Text) != trustedAccessGroup {
		return ErrUnavailable
	}
	groups := values["keychain-access-groups"]
	if groups.XMLName.Local != "array" || len(groups.Children) != 1 {
		return ErrUnavailable
	}
	for _, group := range groups.Children {
		if group.XMLName.Local == "string" && strings.TrimSpace(group.Text) == trustedAccessGroup {
			return nil
		}
	}
	return ErrUnavailable
}

func entitlementValues(raw []byte) (map[string]plistNode, error) {
	if len(raw) > maxMessageBytes {
		return nil, ErrUnavailable
	}
	var root plistNode
	if err := xml.NewDecoder(bytes.NewReader(raw)).Decode(&root); err != nil || root.XMLName.Local != "plist" || len(root.Children) != 1 || root.Children[0].XMLName.Local != "dict" {
		return nil, ErrUnavailable
	}
	children := root.Children[0].Children
	if len(children)%2 != 0 {
		return nil, ErrUnavailable
	}
	result := make(map[string]plistNode)
	for i := 0; i < len(children); i += 2 {
		if children[i].XMLName.Local != "key" {
			return nil, ErrUnavailable
		}
		key := strings.TrimSpace(children[i].Text)
		if _, exists := result[key]; exists {
			return nil, ErrUnavailable
		}
		result[key] = children[i+1]
	}
	return result, nil
}
