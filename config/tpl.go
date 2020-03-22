package config

import (
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

func (z *ZpsConfig) osTest() function.Function {
	return function.New(
		&function.Spec{
			Params: []function.Parameter{
				{
					Name: "value",
					Type: cty.String,
				},
			},
			Type: function.StaticReturnType(cty.String),
			Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
				return args[0], nil
			},
		},
	)
}
