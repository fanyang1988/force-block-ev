package blockdb

import eos "github.com/eosforce/goforceio"

func IsChecksum256Eq(l, r eos.Checksum256) bool {
	if len(l) != len(r) {
		return false
	}

	if len(l) == 0 {
		return true
	}

	for i := 0; i < len(l); i++ {
		if l[i] != r[i] {
			return false
		}
	}

	return true
}

// IsBlockEq return if block l == r
func IsBlockEq(l *eos.SignedBlock, r *eos.SignedBlock) bool {
	lId, _ := l.BlockID()
	rId, _ := r.BlockID()
	// Just check id is ok
	return IsChecksum256Eq(lId, rId) && IsChecksum256Eq(l.Previous, r.Previous)
}
