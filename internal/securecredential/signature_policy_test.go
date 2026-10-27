package securecredential

import (
	"strings"
	"testing"
)

func TestSignatureRejectsInjectedOrDebuggableHelper(t *testing.T) {
	metadata := "Identifier=io.riido.cli-auth-helper\nTeamIdentifier=3PVV8DQPL6\nCodeDirectory v=20500 flags=0x10000(runtime) hashes=5\n"
	safe := `<?xml version="1.0"?><plist><dict><key>application-identifier</key><string>3PVV8DQPL6.io.riido.cli-auth-helper</string><key>keychain-access-groups</key><array><string>3PVV8DQPL6.io.riido.cli-auth-helper</string></array></dict></plist>`
	if err := checkSignature(metadata, []byte(safe)); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"get-task-allow", "com.apple.security.get-task-allow", "com.apple.security.cs.disable-library-validation", "com.apple.security.cs.allow-unsigned-executable-memory", "com.apple.security.cs.allow-dyld-environment-variables", "com.apple.security.cs.allow-jit", "com.apple.security.cs.disable-executable-page-protection"} {
		unsafe := strings.Replace(safe, "</dict>", "<key>"+key+"</key><true/></dict>", 1)
		if checkSignature(metadata, []byte(unsafe)) == nil {
			t.Fatalf("accepted %s", key)
		}
	}
	extra := strings.Replace(safe, "</array>", "<string>3PVV8DQPL6.other</string></array>", 1)
	if checkSignature(metadata, []byte(extra)) == nil {
		t.Fatal("accepted wider access groups")
	}
	if checkSignature("Identifier=io.riido.cli-auth-helper\nTeamIdentifier=3PVV8DQPL6\n", []byte(safe)) == nil {
		t.Fatal("accepted without hardened runtime")
	}
}

func TestSignatureRejectsInvalidEntitlementTypesAndDuplicates(t *testing.T) {
	meta := "Identifier=io.riido.cli-auth-helper\nTeamIdentifier=3PVV8DQPL6\nCodeDirectory flags=0x10000(runtime)\n"
	for _, raw := range []string{`<plist><dict><key>application-identifier</key><string>wrong</string></dict></plist>`, `<plist><dict><key>get-task-allow</key><false/><key>get-task-allow</key><true/></dict></plist>`, `<plist><dict><key>get-task-allow</key><string>false</string></dict></plist>`, "not plist"} {
		if checkSignature(meta, []byte(raw)) == nil {
			t.Fatal("accepted invalid entitlements")
		}
	}
}
