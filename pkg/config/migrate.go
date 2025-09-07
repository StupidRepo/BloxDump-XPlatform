package config

import "reflect"

func MigrateConfig(cfg *Config) {
	defaults := CreateConfig()

	if cfg.Version != defaults.Version {
		vCfg := reflect.ValueOf(cfg).Elem()
		vDef := reflect.ValueOf(defaults)

		for i := 0; i < vCfg.NumField(); i++ {
			field := vCfg.Field(i)
			if field.IsNil() {
				field.Set(vDef.Field(i))
			}
		}

		newMigrations := append(*cfg.Migrations, *cfg.Version)
		cfg.Migrations = &newMigrations
		cfg.Version = defaults.Version
	}
}
