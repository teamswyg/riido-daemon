package saasplane

func New(cfg Config) (*Plane, error) {
	cfg, client := defaultConfig(cfg)
	if err := validateConfig(cfg); err != nil {
		return nil, err
	}
	p := &Plane{
		cfg:    cfg,
		client: client,
		ops:    make(chan stateOp, 64),
		done:   make(chan struct{}),
	}
	go p.loop()
	return p, nil
}
