package csis3

// TODO: move this whole thing to iaas (?) package and see if creds can be put into a struct or something
func awsCreds(s map[string]string) (string, string, bool) {
	key, keyOk := s["AWS_ACCESS_KEY_ID"]
	secret, secretOk := s["AWS_SECRET_ACCESS_KEY"]
	return key, secret, keyOk && secretOk
}
