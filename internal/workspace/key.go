package workspace

func GetWorkspacePublicKey(root string) (string, error) {
	if WorkspaceFileExists(root) {
		wc, err := ReadWorkspaceFile(root)
		if err == nil && wc.PublicKey != "" {
			return wc.PublicKey, nil
		}
	}
	return "", nil
}
