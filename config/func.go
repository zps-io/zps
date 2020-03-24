package config

import (
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

func (z *ZpsConfig) configDefault() function.Function {
	return function.New(
		&function.Spec{
			Params: []function.Parameter{
				{
					Name: "value",
					Type: cty.DynamicPseudoType,
				},
				{
					Name: "default",
					Type: cty.String,
				},
			},
			Type: function.StaticReturnType(cty.String),
			Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
				if args[0].IsNull() {
					return args[1], nil
				}

				return args[0], nil
			},
		},
	)
}
