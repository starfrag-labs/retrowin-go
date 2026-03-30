package etc

import (
	"bytes"
	"slices"
	"strconv"
	"strings"
)

// GroupEntry represents a parsed line from /etc/group.
// Format: group_name:password:gid:member_uids
type GroupEntry struct {
	Name    string
	GID     int
	Members []int
}

// ParseGroupFile parses /etc/group content into GroupEntry slice.
func ParseGroupFile(data []byte) ([]GroupEntry, error) {
	lines := bytes.Split(data, []byte{'\n'})
	var entries []GroupEntry

	for _, line := range lines {
		line = bytes.TrimSpace(line)
		if len(line) == 0 || line[0] == '#' {
			continue
		}

		parts := strings.SplitN(string(line), ":", 4)
		if len(parts) < 3 {
			continue
		}

		gid, err := strconv.Atoi(parts[2])
		if err != nil {
			continue
		}

		entry := GroupEntry{
			Name: parts[0],
			GID:  gid,
		}

		if len(parts) == 4 && parts[3] != "" {
			for m := range strings.SplitSeq(parts[3], ",") {
				m = strings.TrimSpace(m)
				if m == "" {
					continue
				}
				uid, err := strconv.Atoi(m)
				if err != nil {
					continue
				}
				entry.Members = append(entry.Members, uid)
			}
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

// ResolveGIDsByUID returns all gids that the given uid belongs to.
func ResolveGIDsByUID(data []byte, uid int) []int {
	entries, err := ParseGroupFile(data)
	if err != nil {
		return nil
	}

	var gids []int
	for _, entry := range entries {
		if slices.Contains(entry.Members, uid) {
			gids = append(gids, entry.GID)
		}
	}

	return gids
}
