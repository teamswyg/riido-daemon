package supervisor

import "github.com/teamswyg/riido-daemon/pkg/lifecycle"

func New(cfg Config) (*Actor, error) {
	cfg, err := normalizeConfig(cfg)
	if err != nil {
		return nil, err
	}
	return &Actor{
		cfg:       cfg,
		mailbox:   make(chan envelope, cfg.MailboxSize),
		stopReqCh: make(chan lifecycle.ShutdownLevel, cfg.MailboxSize),
		stoppedCh: make(chan struct{}),
		stopErrCh: make(chan error, 1),
	}, nil
}
