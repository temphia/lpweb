package seekers

var salt = "salty1"

func init() {
	_salt := os.GetEnv("LPWEB_SALT")
	if _salt != "" {
		salt = _salt
	}
}

type SeekAddr struct {
	Addrs []string
	Port  string
}

func Hash(pubkey string) (string, error) {

	return "", nil
}

func Encode(pubkey string, addr *SeekAddr) (string, error) {

	return "", nil
}

func Decode(pubkey string, coded string) (*SeekAddr, error) {

	return nil, nil
}
