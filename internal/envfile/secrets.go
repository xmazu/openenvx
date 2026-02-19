package envfile

import (
	"github.com/xmazu/openenvx/internal/scanner"
)

type CommentedSecret struct {
	Line          *Line
	MightBeSecret bool
}

func (f *File) DetectCommentedSecrets(detector *NoEncryptDetector, patterns []scanner.Pattern) []CommentedSecret {
	var results []CommentedSecret

	if detector == nil {
		return []CommentedSecret{}
	}

	for _, line := range f.lines {
		if line.Type != LineTypeCommentedAssignment {
			continue
		}

		var mightBeSecret bool

		if detector.ShouldSkip(line.Key, line.Value) {
			mightBeSecret = false
		} else {
			mightBeSecret = true
		}

		results = append(results, CommentedSecret{
			Line:          line,
			MightBeSecret: mightBeSecret,
		})
	}

	return results
}

func FilterSecrets(secrets []CommentedSecret) []CommentedSecret {
	var filtered []CommentedSecret
	for _, s := range secrets {
		if s.MightBeSecret {
			filtered = append(filtered, s)
		}
	}
	return filtered
}
