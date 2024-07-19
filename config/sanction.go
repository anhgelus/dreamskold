package config

import (
	_ "embed"
	"github.com/anhgelus/dreamskold/sanction"
)

//go:embed sanction.toml
var DefaultSanctionConfig string

type SanctionConfig struct {
	Retentions         []*RetentionConfig `toml:"retentions"`
	BanCommandSanction string             `toml:"ban_command_sanction"`
	Sanctions          []*sanction.Type   `toml:"sanctions"`
}

type RetentionConfig struct {
	ID                int   `toml:"id"`
	Retention         int64 `toml:"retention"`
	RelativeRetention int8  `toml:"relative_retention"`
}
