package seekers

type PubEtcd struct {
}

func (p *PubEtcd) Set(hash, addr string) error {
	return nil
}

func (p *PubEtcd) Get(hash string) (string, error) {
	return "", nil
}
