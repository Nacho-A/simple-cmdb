package casbinx

import (
	"fmt"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/util"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"gorm.io/gorm"
)

var Enforcer *casbin.Enforcer

func Init(modelConfPath string, db *gorm.DB) (*casbin.Enforcer, error) {
	m, err := model.NewModelFromFile(modelConfPath)
	if err != nil {
		return nil, fmt.Errorf("load casbin model: %w", err)
	}

	adapter, err := gormadapter.NewAdapterByDB(db)
	if err != nil {
		return nil, fmt.Errorf("init casbin adapter: %w", err)
	}

	e, err := casbin.NewEnforcer(m, adapter)
	if err != nil {
		return nil, fmt.Errorf("new enforcer: %w", err)
	}

	e.AddFunction("keyMatch2", func(args ...interface{}) (interface{}, error) {
		if len(args) != 2 {
			return false, nil
		}
		k1, ok1 := args[0].(string)
		k2, ok2 := args[1].(string)
		if !ok1 || !ok2 {
			return false, nil
		}
		return util.KeyMatch2(k1, k2), nil
	})
	if err := e.LoadPolicy(); err != nil {
		return nil, fmt.Errorf("load policy: %w", err)
	}

	Enforcer = e
	return e, nil
}

