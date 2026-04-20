package scheduler

import (
	"strings"
)

// TruncateString truncates a string to a maximum number of runes.
// It ensures that UTF-8 characters are not split and attempts to fix simple Markdown bold tags.
func TruncateString(s string, maxRunes int) string {
	runes := []rune(s)
	if len(runes) <= maxRunes {
		return s
	}

	truncatedRunes := runes[:maxRunes]
	truncated := string(truncatedRunes)

	// Fix broken bold tags at the very end
	// Case 1: "word *" -> Remove the trailing "*"
	if strings.HasSuffix(truncated, "*") && !strings.HasSuffix(truncated, "**") {
		truncated = truncated[:len(truncated)-1]
	}
	// Case 2: "word **" -> This is balanced if there's an opening "**"
	// We need to check the TOTAL count of "**" in the truncated string.
	
	boldCount := strings.Count(truncated, "**")
	if boldCount%2 != 0 {
		// If odd, we have an open bold tag. Close it.
		// BUT, if it ALREADY ends with "**", it means the LAST tag is what's making it odd.
		// E.g., "text **bold" -> odd, append "**" -> "text **bold**..." (Correct)
		// E.g., "text **" -> odd, appending "**" would make it "text ****..." (Wrong)
		// So if it ends with "**", we should probably just remove it or handle it.
		if strings.HasSuffix(truncated, "**") {
			// It ends with an opening tag. Let's just remove the opening tag to be clean.
			truncated = strings.TrimSuffix(truncated, "**")
		} else {
			truncated += "**"
		}
	}

	return strings.TrimSpace(truncated) + "..."
}

// GetMentionTag returns the proper Discord mention string based on the provided input.
// If it's "everyone" or "here", it returns "@everyone" or "@here".
// If it's a numeric ID, it returns "<@&ID>".
func GetMentionTag(mentionRoleID string) string {
	if mentionRoleID == "" {
		return ""
	}
	if mentionRoleID == "everyone" || mentionRoleID == "here" {
		return "@" + mentionRoleID
	}
	return "<@&" + mentionRoleID + ">"
}
