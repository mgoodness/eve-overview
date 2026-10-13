package overview

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
)

// groupMatrix is the SDE's Group-level lookup: the name of a Group to the
// identifier (or identifiers, when names collide) the preset whitelist needs.
type groupMatrix struct {
	idsByName map[string][]int
}

// parseGroupMatrix reads the SDE's groups.jsonl: one JSON object per line,
// each with a `_key` identifier and a `name.en` localized name.
func parseGroupMatrix(jsonl []byte) (*groupMatrix, error) {
	matrix := &groupMatrix{idsByName: make(map[string][]int)}
	scanner := bufio.NewScanner(bytes.NewReader(jsonl))
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		var group struct {
			Key  int `json:"_key"`
			Name struct {
				En string `json:"en"`
			} `json:"name"`
		}
		if err := json.Unmarshal(line, &group); err != nil {
			return nil, fmt.Errorf("parse SDE Group matrix: %w", err)
		}
		matrix.idsByName[group.Name.En] = append(matrix.idsByName[group.Name.En], group.Key)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read SDE Group matrix: %w", err)
	}
	return matrix, nil
}

// resolve maps a Group name to exactly one identifier. A name the SDE does not
// know, or a name shared by two Groups, is an error naming the Group: inside a
// preset the whitelist gates before any State can override it, so a silently
// dropped Group is unrecoverable.
func (m *groupMatrix) resolve(name string) (int, error) {
	ids := m.idsByName[name]
	switch len(ids) {
	case 0:
		return 0, fmt.Errorf("group %q does not resolve in the SDE matrix", name)
	case 1:
		return ids[0], nil
	default:
		return 0, fmt.Errorf("group %q is ambiguous in the SDE matrix (identifiers %v)", name, ids)
	}
}
